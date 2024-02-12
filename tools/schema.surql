-- define users table
DEFINE TABLE user SCHEMAFULL;

DEFINE FIELD name ON user TYPE string;

DEFINE FIELD email ON user TYPE string ASSERT string::is::email($value);
DEFINE INDEX emailIndex ON TABLE user COLUMNS email UNIQUE;

DEFINE FIELD impartus_password ON user TYPE option<string>;
DEFINE FIELD cms_token ON user TYPE option<string>;

-- define impartus auth tokens table
DELETE impartus_token;
DEFINE TABLE impartus_token AS SELECT 
    id as user,
    time::now() as updated_at
FROM user WHERE type::is::string(impartus_password);

# token is only set when the password is valid
DEFINE FIELD token ON impartus_token TYPE option<string> DEFAULT NONE;