# main.py

import sys
import argparse
import traceback
import asyncio
from fastapi import FastAPI # type: ignore
from fastapi.responses import RedirectResponse # type: ignore
import uvicorn # type: ignore

from server.routes import router
from server.services.db_service import db_service
from server.services.session_service import session_service
from server.services.killswitch_service import KillSwitchService
from server.middleware import ActivityMiddleware
from server.config import settings

def main():
    print("[DEBUG] Starting Presenz backend...")

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
    print(f"[DEBUG] Using DB: {db_path}")

    # -------------------------
    # Initialize DB
    # -------------------------
    try:
        db_service.connect(db_path)
        print("[DEBUG] DB connection established")
    except Exception:
        print("[ERROR] Failed to connect to DB")
        traceback.print_exc()
        sys.exit(1)

    # -------------------------
    # Initialize session
    # -------------------------
    try:
        # new (correct)
        session_service.start_session(
            max_count=args.total,
            course=args.course,
            batch=args.batch,
            db_filename=db_path.split("/")[-1],  # just the filename
        )
        table_name = session_service.get_table_name
        session_code = session_service.get_session_code
        print("+------------------------------------------------------------------------------------+")
        print(f" [DEBUG] Session initialized: {table_name}")
        print(f" [DEBUG] Session code (share with students): {session_code}")
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
        print(f"[DEBUG] Attendance table created: {table_name}")
    except Exception:
        print("[ERROR] Failed to create attendance table")
        traceback.print_exc()
        sys.exit(1)

    # -------------------------
    # Launch FastAPI app
    # -------------------------
    app = FastAPI(title="Presenz Attendance System")
    @app.get("/")
    def root():
        return RedirectResponse(url="/attendance/")
    app.include_router(router, prefix="/attendance")

    # -------------------------
    # Initialize KillSwitch
    # -------------------------
    killswitch = KillSwitchService()
    print("[DEBUG] FastAPI app initialized")
    print("[DEBUG] Presenz is ready to accept attendance submissions")

    # -------------------------
    # Run server with kill switch
    # -------------------------
    async def run_server():
        config = uvicorn.Config(app, host=settings.server_host, port=settings.server_port)
        server = uvicorn.Server(config)

        # -------------------------
        # Launch server and KillSwitch tasks concurrently
        # -------------------------
        server_task = asyncio.create_task(server.serve())
        # monitor_task = asyncio.create_task(killswitch.inactivity_monitor())
        listener_task = asyncio.create_task(killswitch.manual_terminate_listener())

        try:
            # Wait until KillSwitch triggers shutdown
            await killswitch.wait_for_shutdown()
            print("[DEBUG] KillSwitch triggered shutdown.")
            server.should_exit = True  # Graceful shutdown

            # Wait for the uvicorn server to exit cleanly
            await server_task

        except Exception as e:
            print("[ERROR] Exception in server run:", e)

        finally:
            # Cleanup DB and session
            db_service.close()
            session_service.end_session()
            print("[DEBUG] Server shutdown gracefully.")

            # Cancel background tasks if still running
            for task in [listener_task]:
                if not task.done():
                    task.cancel()
                    try:
                        await task
                    except asyncio.CancelledError:
                        pass

    # -------------------------
    # Launch server
    # -------------------------
    try:
        asyncio.run(run_server())
    except Exception:
        print("[ERROR] Failed to start FastAPI server")
        traceback.print_exc()
        sys.exit(1)

if __name__ == "__main__":
    main()