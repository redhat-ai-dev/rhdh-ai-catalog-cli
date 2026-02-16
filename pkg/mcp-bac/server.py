import anyio
import click
import subprocess
import mcp.types as types
from mcp.server.lowlevel import Server


async def fetch_models(
        url: str,
        token: str,
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    try:
        command = 'bac get components --backstage-url=%s --backstage-token=%s --backstage-skip-tls=true ' % (url, token)
        process = subprocess.run(command, capture_output=True, text=True, check=False)
        stdout = process.stdout.strip()
        stderr = process.stderr.strip()
        if len(stdout) > 0:
           return [types.TextContent(type="text", text=stdout)]
        return [types.TextContent(type="text", text=stderr)]
    except FileNotFoundError:
        return {"stderr": f"Error: Command not found: bac", "returncode": 127}
    except Exception as e:
        return {"stderr": f"An unexpected error occurred: {e}", "returncode": 1}



@click.command()
@click.option("--port", default=8000, help="Port to listen on for SSE")
@click.option(
    "--transport",
    type=click.Choice(["stdio", "sse"]),
    default="stdio",
    help="Transport type",
)
def main(port: int, transport: str) -> int:
    app = Server("mcp-backstage-fetcher")

    @app.call_tool()
    async def fetch_tool(
            name: str, arguments: dict
    ) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
        if name != "fetch":
            raise ValueError(f"Unknown tool: {name}")
        if "url" not in arguments:
            raise ValueError("Missing required argument 'url'")
        if "token" not in arguments:
            raise ValueError("Missing required argument 'token'")
        return await fetch_models(arguments["url", "token"])

    @app.list_tools()
    async def list_tools() -> list[types.Tool]:
        return [
            types.Tool(
                name="fetch-models",
                description="Fetches AI models imported into a backstage instance",
                inputSchema={
                    "type": "object",
                    "required": ["url"],
                    "properties": {
                        "url": {
                            "type": "string",
                            "description": "URL for backstage",
                        },
                        "token": {
                            "type": "string",
                            "description": "token for backstage",
                        }
                    },
                },
            )
        ]

    if transport == "sse":
        from mcp.server.sse import SseServerTransport
        from starlette.applications import Starlette
        from starlette.responses import Response
        from starlette.routing import Mount, Route

        sse = SseServerTransport("/messages/")

        async def handle_sse(request):
            async with sse.connect_sse(
                    request.scope, request.receive, request._send
            ) as streams:
                await app.run(
                    streams[0], streams[1], app.create_initialization_options()
                )
            return Response()

        starlette_app = Starlette(
            debug=True,
            routes=[
                Route("/sse", endpoint=handle_sse, methods=["GET"]),
                Mount("/messages/", app=sse.handle_post_message),
            ],
        )

        import uvicorn

        uvicorn.run(starlette_app, host="0.0.0.0", port=port)
    else:
        from mcp.server.stdio import stdio_server

        async def arun():
            async with stdio_server() as streams:
                await app.run(
                    streams[0], streams[1], app.create_initialization_options()
                )

        anyio.run(arun)

    return 0