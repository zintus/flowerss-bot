# Instructions for Claude

## Code Changes

All code changes must be made using the `mcp__aider_mcp__launch_code_editing_agent` tool. Do not use Edit, MultiEdit, or Write tools for modifying code files. Instead, use the aider_mcp agent with clear, high-level instructions about what needs to be changed.

## Project Information

This is a Telegram bot project written in Go that monitors RSS feeds and sends updates to users.

## Testing

To test the project, use:
```bash
go test ./...
```

## Building

To build the project, use:
```bash
go build
```