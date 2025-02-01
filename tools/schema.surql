// Remove the testing namespace
// need to find a better way to do this
REMOVE NAMESPACE IF EXISTS lex_test;

USE NS lex;

-- ------
-- IMPARTUS
-- ------
USE DB impartus;


DEFINE TABLE user SCHEMAFULL;
DEFINE FIELD email ON user TYPE string ASSERT string::is::email($value);
DEFINE FIELD password ON user TYPE string;
DEFINE FIELD jwt ON user TYPE option<string>;
DEFINE FIELD updated_at ON user TYPE datetime DEFAULT time::now();

DEFINE INDEX userEmailIndex ON TABLE user COLUMNS email UNIQUE;

DEFINE FUNCTION OVERWRITE fn::get_token($user: record<user>) {
    // return none if the last password fetch attempt failed
    IF type::is::none($user.jwt) {
        RETURN NONE;
    };

    IF (time::now() - $user.updated_at > 6d) {
        UPDATE ONLY $user SET jwt = function() {
            const response = await fetch("https://bitshyd.impartus.com/api/auth/signin", {
                method: "POST",
                body: JSON.stringify({
                    "username": this.email,
                    "password": this.password,
                }),
                headers: {
                    "content-type": "application/json",
                    // need to set user agent for impartus api
                    "user-agent": "axios"
                }
            });

            if (response.status != 200) return undefined;

            const data = await response.json();
            return data.token;
        };
    };
    
    RETURN $user.jwt;
};

// update timestamp when the jwt field has been changed
DEFINE EVENT token_updated ON user 
    WHEN (($before.jwt != $after.jwt) && $event = "UPDATE") 
    THEN 
{
    UPDATE ONLY $value SET updated_at = time::now();
};

// Defines the course independently of the session or subject id.
DEFINE TABLE subject SCHEMAFULL;
// id field is of the form `[department (eg: CS,ECE), subject_code (eg, F211)]`
DEFINE FIELD department ON subject TYPE string;
DEFINE FIELD code ON subject TYPE string;
DEFINE FIELD name ON subject TYPE string;

// Defines the lectures across various years and lecture sections.
DEFINE TABLE lecture SCHEMAFULL;
// id field is of the form `[session_id (defines which sem the course was taken), subject_id]`
DEFINE FIELD impartus_session ON lecture TYPE int;
DEFINE FIELD impartus_subject ON lecture TYPE int;
DEFINE FIELD subject ON lecture TYPE record<subject>;
DEFINE FIELD section ON lecture TYPE string;
DEFINE FIELD professor ON lecture TYPE string;
DEFINE FIELD users ON lecture TYPE set<record<user>>;


DEFINE TABLE pinned TYPE RELATION FROM user TO subject ENFORCED;
DEFINE INDEX unique_relationships
    ON TABLE pinned
    COLUMNS in, out UNIQUE;

// Stores the lectures whose names could not be parsed automatically
DEFINE TABLE malformed_lectures SCHEMAFULL;
DEFINE FIELD users ON malformed_lectures TYPE set<record<user>>;
DEFINE FIELD data ON malformed_lectures FLEXIBLE TYPE object;
DEFINE FIELD fixed ON malformed_lectures TYPE bool DEFAULT false;

// Checks if the lecture name contains all the information needed by lex
DEFINE FUNCTION OVERWRITE fn::extra::check_lecture_name($name: string) {
    let $name_array = string::words($name);
    IF array::len($name_array) <= 3 {
        RETURN false;
    };

    let $subject_code = array::at($name_array, 1);
    // subject code should be of the form CXXX (like F123)
    IF !(string::is::alpha(string::slice($subject_code, 0, 1)) && 
        string::is::numeric(string::slice($subject_code, 1))) {
        RETURN false;
    };

    let $lecture_section = array::at($name_array, -1);
    // lecture section should be of the form CX (like L1 or T1)
    IF !(string::is::alpha(string::slice($lecture_section, 0, 1)) && 
        string::is::numeric(string::slice($lecture_section, 1))) {
        RETURN false;
    };

    RETURN true;
};

// Used to extract the registered lectures and add it to the global dataset
DEFINE FUNCTION OVERWRITE fn::extract_lectures($user: record<user>) {
    let $impartus_jwt = fn::get_token($user);
    IF type::is::none($impartus_jwt) {
        THROW "can't get user's impartus jwt";
    };

    let $lectures = http::get("https://bitshyd.impartus.com/api/subjects", {
        "Authorization": string::concat("Bearer ", $impartus_jwt),
        "user-agent": "axios"
    });

    FOR $lecture IN $lectures {
        IF !fn::extra::check_lecture_name($lecture.subjectName) {
            // extract malformed lecture
            UPSERT type::thing("malformed_lectures", $lecture.subjectName) SET users += $user, data = $lecture;
            CONTINUE;
        };

        // should be like ["BITS", "F111", "SIGMA", "BOY", "L1"]
        let $name_array = string::words($lecture.subjectName);

        let $department = array::at($name_array, 0);
        let $subject_code = array::at($name_array, 1);
        let $lecture_section = array::at($name_array, -1);

        // contains just the subject name, like "OPERATING SYSTEMS"
        let $subject_name = array::join(array::slice($name_array, 2, -1), " ");

        // Create the subjects
        UPSERT subject:[$department, $subject_code] CONTENT {
            department: $department,
            code: $subject_code,
            name: $subject_name
        };

        let $impartus_session = type::int($lecture.sessionId);
        let $impartus_subject = type::int($lecture.subjectId);

        // create a lecture and add user registered to lecture
        IF record::exists(lecture:[$impartus_session, $impartus_subject]) {
            UPDATE lecture:[$impartus_session, $impartus_subject] SET users += $user;
        } ELSE {
            CREATE lecture:[$impartus_session, $impartus_subject] CONTENT {
                impartus_session: $impartus_session,
                impartus_subject: $impartus_subject,
                subject: subject:[$department, $subject_code],
                section: $lecture_section,
                professor: string::trim($lecture.professorName),
                users: [$user]
            };
        };
    };

    RETURN count(SELECT id FROM lecture WHERE users ?= $user);
};

// Used to pin the registered subjects of the latest session 
DEFINE FUNCTION OVERWRITE fn::pin_registered($user: record<user>) {
    let $impartus_jwt = fn::get_token($user);
    IF type::is::none($impartus_jwt) {
        THROW "can't get user's impartus jwt";
    };
    
    let $lectures = http::get("https://bitshyd.impartus.com/api/subjects", {
        'Authorization': string::concat("Bearer ", $impartus_jwt)
    });

    // this session id is for the one the student is currently part of
    let $latest_session = $lectures[0].sessionId;
    FOR $lecture IN $lectures[WHERE sessionId = $latest_session] {
        IF !fn::extra::check_lecture_name($lecture.subjectName) {
            CONTINUE;
        };

        let $name_array = string::words($lecture.subjectName);
        let $department = array::at($name_array, 0);
        let $subject_code = array::at($name_array, 1);

        
        IF array::is_empty(SELECT value id FROM pinned WHERE in = $user AND out = subject:[$department, $subject_code]) {
            RELATE $user->pinned->subject:[$department, $subject_code] RETURN NONE;
        };
    };

    RETURN count(SELECT id FROM pinned WHERE in = $user);
};

DEFINE ANALYZER OVERWRITE subject_analyzer TOKENIZERS blank FILTERS ascii,lowercase,edgengram(2,10);
DEFINE INDEX OVERWRITE subject_search ON subject FIELDS name,department,code SEARCH ANALYZER subject_analyzer BM25;