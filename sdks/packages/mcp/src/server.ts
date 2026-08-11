/** Build a configured McpServer instance with all curated go-ios tools registered. */
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { GoIosClient } from "./client.js";
import type { ServerConfig } from "./config.js";
import { registerTools } from "./tools.js";

export const SERVER_NAME = "go-ios-mcp";
export const SERVER_VERSION = "0.1.0";

export function createServer(config: ServerConfig): McpServer {
  const server = new McpServer(
    { name: SERVER_NAME, version: SERVER_VERSION },
    {
      instructions:
        "Control iOS devices through the go-ios daemon. Start with list_devices to get a udid; " +
        "every other tool is device-scoped and needs that udid. Tools by area: info/health " +
        "(device_info, device_health, device_battery, device_diagnostics, device_diskspace, " +
        "device_ip, list_processes); apps (list_apps, launch_app, terminate_app, install_app, " +
        "uninstall_app); screen & logs (screenshot, stream_logs); performance " +
        "(sample_performance); crash logs (list_crash_reports, pull_crash_report); files, " +
        "read-only (list_files, list_device_files, read_file); location (set_location_gpx); " +
        "pasteboard (get_pasteboard, set_pasteboard); web debugging (list_webinspector_pages, " +
        "webinspector_eval). To DRIVE the UI: first boot WebDriverAgent — run_wda (or " +
        "create_wda_session) — and forward its port, then act with the ui_* tools (ui_tap, " +
        "ui_swipe, ui_type, ui_press_button, ui_app_launch, ui_app_terminate) and observe with " +
        "ui_source (view hierarchy) and ui_screenshot; pass the forwarded WDA base URL as wdaUrl. " +
        "Track long jobs with list_jobs, get_job, tail_job_logs, stop_job. Device management is " +
        "DISRUPTIVE (reboot_device, shutdown_device). stream_logs, sample_performance and " +
        "tail_job_logs are bounded captures that always return within their duration/line caps.",
    },
  );
  const client = new GoIosClient(config);
  registerTools(server, client);
  return server;
}
