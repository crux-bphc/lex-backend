from surrealdb import Surreal
import typing
import aiohttp
import os

REVALIDATION_PERIOD = "6d"  # see: https://docs.surrealdb.com/docs/surrealql/datamodel/datetimes#duration-units

IMPARTUS_URL = "http://a.impartus.com/api/{endpoint}"


# typed dicts for convenience
class ImpartusUser(typing.TypedDict):
    email: str
    impartus_password: str


class ImpartusToken(typing.TypedDict):
    id: str
    token: str | None
    user: ImpartusUser


async def revalidate_impartus_tokens():
    """
    Revalidates old impartus auth tokens.

    Selects all tokens from the table *impartus_token* which either have no *token* field
    or *updated_at* timestamp is older than 6 days.

    If it satisfies this condition then fetch new token based on the user's *email* and
    *impartus_password*.

    If for some reason the fetch fails or the password is incorrect then set the *token*
    field as *NONE*.
    """
    async with Surreal("ws://db:8000/rpc") as db:
        await db.signin(
            {"user": os.environ["DB_USER"], "pass": os.environ["DB_PASSWORD"]}
        )
        await db.use(os.environ["DB_NAMESPACE"], os.environ["DB_DATABASE"])

        tokens = await _query_invalid_tokens(db)

        async with aiohttp.ClientSession() as session:
            l = await asyncio.gather(
                *(_fetch_new_token(session, token) for token in tokens)
            )

        await _update_db_tokens(db, l)

        success_count = sum(token[1] is not None for token in l)
        print(f"Updated {success_count}/{len(tokens)} users successfully")


async def _query_invalid_tokens(db: Surreal) -> list[ImpartusToken]:
    results = await db.query(
        f"""
			SELECT id, token, user.impartus_password, user.email
			FROM impartus_token
			WHERE (token == NONE) OR (time::now() - updated_at > {REVALIDATION_PERIOD});
		"""
    )
    # return the result of the first (and only) query
    return results[0]["result"]


async def _fetch_new_token(
    session: aiohttp.ClientSession, token: ImpartusToken
) -> tuple[str, str | None]:
    """
    Returns (database record id, new token) as a tuple.

    If there was an error fetching the new token, it is set to `None`
    """
    body = {
        "username": token["user"]["email"],
        "password": token["user"]["impartus_password"],
    }

    # post request to signin endpoint
    async with session.post(
        IMPARTUS_URL.format(endpoint="auth/signin"), data=body
    ) as r:
        # check for success
        if r.status == 200:
            j = await r.json()
            return (token["id"], j["token"])
        else:
            print(
                f"Could not fetch token for user: {token['user']['email']}, {(await r.read()).decode()}"
            )
            return (token["id"], None)


async def _update_db_tokens(db: Surreal, tokens: list[tuple[str, str | None]]):
    for token in tokens:
        placeholder = "NONE" if token[1] is None else "$value"
        await db.query(
            f"UPDATE $user_token SET token = {placeholder}, updated_at = time::now()",
            {"user_token": token[0], "value": token[1]},
        )


if __name__ == "__main__":
    import asyncio

    asyncio.run(revalidate_impartus_tokens())