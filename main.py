# main.py

import sys
import argparse
import traceback
import asyncio
from fastapi import FastAPI # type: ignore
from fastapi.responses import RedirectResponse # type: ignore
from slowapi import _rate_limit_exceeded_handler
from slowapi.errors import RateLimitExceeded
import uvicorn # type: ignore

from server.routes import router
from server.routes.attendance import limiter
from server.services.db_service import db_service
from server.services.session_service import session_service
from server.services.killswitch_service import KillSwitchService
from server.middleware import ActivityMiddleware
from server.config import settings

def main():
    print("\n[Presenz] Starting Presenz backend...")

    # -------------------------
    # Parse CLI arguments
    # -------------------------
    parser = argparse.ArgumentParser(description="Presenz attendance system")
    parser.add_argument("--course", required=True, help="Course name")
    parser.add_argument("--batch", required=True, help="Batch ID")
    parser.add_argument("--total", required=True, type=int, help="Total number of students")
    parser.add_argument("--db", required=False, help="SQLite DB file path")
    args = parser.parse_args()

    db_path = args.db if args.db else settings.default_db
    print(f"[Presenz] Using DB: {db_path}")

    # -------------------------
    # Initialize DB
    # -------------------------
    try:
        db_service.connect(db_path)
        print("[Presenz] DB connection established")
    except Exception:
        print("[ERROR] Failed to connect to DB")
        traceback.print_exc()
        sys.exit(1)

    # -------------------------
    # Initialize session
    # -------------------------
    try:
        session_service.start_session(
            max_count=args.total,
            course=args.course,
            batch=args.batch,
            db_filename=db_path.split("/")[-1],
        )
        table_name = session_service.get_table_name
        session_code = session_service.get_session_code
        print("+------------------------------------------------------------------------------------+")
        print(f" [Presenz] Session initialized: {table_name}")
        print(f" [Presenz] Session code (share with students): {session_code}")
        print("+------------------------------------------------------------------------------------+")
    except Exception:
        print("[ERROR] Failed to initialize session")
        traceback.print_exc()
        sys.exit(1)

    # -------------------------
    # Create attendance table
    # -------------------------
    try:
        db_service.create_table(table_name)
        print(f"[Presenz] Attendance table created: {table_name}")
    except Exception:
        print("[ERROR] Failed to create attendance table")
        traceback.print_exc()
        sys.exit(1)

    # -------------------------
    # Run server with kill switch
    # -------------------------
    async def run_server(app, killswitch):
        config = uvicorn.Config(app, host=settings.server_host, port=settings.server_port)
        server = uvicorn.Server(config)

        server_task = asyncio.create_task(server.serve())
        monitor_task = asyncio.create_task(killswitch.inactivity_monitor())
        listener_task = asyncio.create_task(killswitch.manual_terminate_listener())

        try:
            await killswitch.wait_for_shutdown()
            print("[Presenz] KillSwitch triggered shutdown.")
            server.should_exit = True
            await server_task

        except Exception as e:
            print("[ERROR] Exception in server run:", e)

        finally:
            print("[DEBUG] Server Halted gracefully.\n")

            for task in [listener_task, monitor_task]:
                if not task.done():
                    task.cancel()
                    try:
                        await task
                    except asyncio.CancelledError:
                        pass

    # -------------------------
    # Launch server
    # -------------------------
    print("[Presenz] FastAPI app initialized")
    print("[Presenz] Presenz is ready to accept attendance submissions")

    while True:
        try:
            killswitch = KillSwitchService()
            app = FastAPI(title="Presenz Attendance System")
            app.state.limiter = limiter
            app.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)
            @app.get("/")
            def root():
                return RedirectResponse(url="/attendance/")
            app.include_router(router, prefix="/attendance")
            app.add_middleware(ActivityMiddleware, killswitch=killswitch)

            asyncio.run(run_server(app, killswitch))
            break
        except KeyboardInterrupt:
            print("\n[Presenz] Ctrl+C detected. Do you want to quit? [y/N]: ", end="", flush=True)
            try:
                answer = input().strip().lower()
            except (EOFError, OSError):
                answer = "n"
            if answer == "y":
                db_service.close()
                session_service.end_session()
                print("[Presenz] Shutting down...\n")
                break
            print("[Presenz] Resuming...")
        except Exception:
            print("[ERROR] Failed to start FastAPI server")
            traceback.print_exc()
            sys.exit(1)

if __name__ == "__main__":
    main()