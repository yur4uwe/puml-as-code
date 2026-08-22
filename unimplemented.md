# Unimplemented Features & Backlog

This file tracks advanced capabilities and security hardening for diagram resolution that are desirable for future production/multi-tenant environments, but intentionally omitted from the MVP.

---

## Filesystem Sandboxing & Symlink Containment
- Restricting `!include` paths so traversal (`../`) cannot escape the project root.
- Validating targets with `filepath.EvalSymlinks` to ensure symlinks inside the workspace do not point to sensitive host files.

## Standard Library Virtual File System (`<...>`)
- Create a virtual file system that includes the standard library (`<...>`)

## Remote URL Includes (`http://`, `https://`)
- Add support for HTTP/HTTPS remote URLs (`http://`, `https://`) with a new function inside FileReader interface: ReadRemoteFile(uri string) ([]byte, error)

