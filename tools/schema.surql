// Remove the testing namespace
// need to find a better way to do this
REMOVE NAMESPACE IF EXISTS lex_test;

-- ------
-- IMPARTUS
-- ------
USE DB impartus;

-- define the users table
DEFINE TABLE user SCHEMAFULL;

DEFINE FIELD email ON user TYPE string ASSERT string::is::email($value);
DEFINE FIELD password ON user TYPE string;
DEFINE FIELD jwt ON user TYPE option<string>;
DEFINE FIELD updated_at ON user TYPE datetime DEFAULT time::now();

DEFINE INDEX userEmailIndex ON TABLE user COLUMNS email UNIQUE;

DEFINE FUNCTION fn::get_token($user: record<user>) {
    // return none if the last password fetch attempt failed
    IF (type::is::none($user.jwt)) THEN
        RETURN NONE;
    END;

    IF (time::now() - $user.updated_at > 6d) THEN
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
    END;
    
    RETURN $user.jwt;
};

// 
DEFINE EVENT token_updated ON user 
    WHEN (($before.jwt != $after.jwt) && $event = "UPDATE") 
    THEN 
{
    UPDATE ONLY $value.id SET updated_at = time::now();
};

-- define the subjects table
// Defines the course independently of the session or subject id.
DEFINE TABLE subject SCHEMAFULL;
// id field is of the form `[department (eg: CS,ECE), subject_code (eg, F211)]`
DEFINE FIELD name ON subject TYPE string;

-- define the lectures table
// Defines the lectures across various years and lecture sections.
DEFINE TABLE lecture SCHEMAFULL;
// id field is of the form `[session_id (defines which sem the course was taken), subject_id]`
DEFINE FIELD subject ON lecture TYPE record<subject>;
DEFINE FIELD section ON lecture TYPE int;
DEFINE FIELD professor ON lecture TYPE string;
DEFINE FIELD users ON lecture TYPE set<record<user>>;

-- define the pinned lectures table
DEFINE TABLE pinned TYPE RELATION FROM user TO subject;

// Used to extract the registered lectures and add it to the global dataset
DEFINE FUNCTION fn::extract_lectures($user: record<user>) {
    let $impartus_jwt = fn::get_token($user);
    let $lectures = http::get("https://bitshyd.impartus.com/api/subjects", {
        'Authorization': string::concat("Bearer ", $impartus_jwt)
    });

    FOR $lecture IN $lectures {
        let $name_array = string::words($lecture.subjectName);
        let $department = array::at($name_array, 0);
        let $subject_code = array::at($name_array, 1);
        let $subject_name = array::join(array::slice($name_array, 2, -1), " ");

        // Insert the subjects
        UPDATE subject:[$department, $subject_code] CONTENT {
            name: $subject_name
        } RETURN NONE;

        // Insert the lectures
        UPDATE lecture:[$lecture.sessionId, $lecture.subjectId] SET 
            subject = subject:[$department, $subject_code],
            section = <int> string::slice($lecture.subjectName, -1, 1),
            professor = string::trim($lecture.professorName),
            users += $user
        RETURN NONE;
    };

    RETURN NONE;
};

// Used to pin the registered subjects of the latest session 
DEFINE FUNCTION fn::pin_registered($user: record<user>) {
    let $impartus_jwt = fn::get_token($user);
    let $lectures = http::get("https://bitshyd.impartus.com/api/subjects", {
        'Authorization': string::concat("Bearer ", $impartus_jwt)
    });

    // this session id is for the one the student is currently part of
    let $latest_session = $lectures[0].sessionId;
    FOR $lecture IN $lectures[WHERE sessionId = $latest_session] {
        let $name_array = string::words($lecture.subjectName);
        let $department = array::at($name_array, 0);
        let $subject_code = array::at($name_array, 1);

        RELATE $user->pinned->subject:[$department, $subject_code] RETURN NONE;
    };

    RETURN NONE;
};


DEFINE EVENT extract_on_create ON user 
    WHEN ($event = "CREATE" && type::is::string($value.jwt)) 
    THEN 
{

    // Add the registered lectures of the user to the database
    let $extracted_lectures = fn::extract_lectures($value.id);

    // Pin the registed of the newly created user by default when they register
    let $pinned_subjects = fn::pin_registered($value.id)
};