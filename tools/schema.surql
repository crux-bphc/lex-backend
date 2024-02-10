-- Users table

DEFINE TABLE user SCHEMAFULL;

DEFINE FIELD name ON user TYPE string;

DEFINE FIELD email ON user TYPE string ASSERT string::is::email($value);
DEFINE INDEX userEmailIndex ON TABLE user COLUMNS email UNIQUE;

DEFINE FIELD impartus_password ON user TYPE option<string>;

DEFINE FIELD cms_token ON user TYPE option<string>;