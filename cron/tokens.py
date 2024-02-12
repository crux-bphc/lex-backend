from surrealdb import Surreal
import os

async def revalidate_impartus_tokens():
    """
    TODO: revalidates old impartus auth tokens.

    Selects all tokens from the table *impartus_token* which either have no *token* field
    or *updated_at* timestamp is older than 6 days.

    If it satisfies this condition then fetch new token based on the user's *email* and
    *impartus_password*.
    
    If for some reason the fetch fails or the password is incorrect then set the *token*
    field as *NONE*.
    """
    async with Surreal("ws://db:8000/rpc") as db:
        await db.signin({"user": os.environ["DB_USER"], "pass": os.environ["DB_PASSWORD"]})
        await db.use(os.environ["DB_NAMESPACE"], os.environ["DB_DATABASE"])

if __name__ == "__main__":
    import asyncio
    asyncio.run(revalidate_impartus_tokens())