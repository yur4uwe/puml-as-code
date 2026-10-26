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

## Unclosed Block Heuristic Hinting at First Bad Member
- In structured dialects (e.g. `GoDialect`), when an unclosed block reaches `@enduml` or `EOF`, inspect accumulated members to locate the first `*ast.BadMember`.
- If no valid members follow that bad member (trailing validity check), attach a secondary diagnostic note / hint (`DiagnosticRelatedInformation`): "did you mean to close the block before this line?".
- Deferred because `LaxDialect` is expected to be preferred for informal sketching, where lines are accepted without syntax validation and do not generate `BadMember` nodes.

