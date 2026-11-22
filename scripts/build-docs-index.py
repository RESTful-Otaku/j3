#!/usr/bin/env python3
"""Generate the docs site's API index and search index from the OpenAPI contract.

Run via `bun run sync:api-docs` so the committed static files stay in sync with
api/openapi.yaml and docs/*.md. No third-party Python packages beyond PyYAML.
"""
import json
import re
import sys
from pathlib import Path

import yaml

METHODS = {"get", "post", "put", "patch", "delete"}


def slugify(value: str) -> str:
    value = re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")
    return value


def example_for(schema, components, seen=None, depth=0):
    if not schema or depth > 4:
        return None
    if "$ref" in schema:
        name = schema["$ref"].rsplit("/", 1)[-1]
        seen = set() if seen is None else set(seen)
        if name in seen:
            return {}
        seen.add(name)
        return example_for(components.get("schemas", {}).get(name, {}), components, seen, depth + 1)
    if "example" in schema:
        return schema["example"]
    if "default" in schema:
        return schema["default"]
    if schema.get("enum"):
        return schema["enum"][0]
    kind = schema.get("type")
    if isinstance(kind, list):
        kind = next((item for item in kind if item != "null"), kind[0] if kind else None)
    if kind == "object":
        return {
            key: example_for(value, components, seen, depth + 1)
            for key, value in schema.get("properties", {}).items()
        }
    if kind == "array":
        item = example_for(schema.get("items", {}), components, seen, depth + 1)
        return [item] if item is not None else []
    if kind in ("integer", "number"):
        return 0
    if kind == "boolean":
        return False
    if kind == "string":
        fmt = schema.get("format")
        if fmt == "date-time":
            return "2026-11-22T10:00:00Z"
        if fmt == "date":
            return "2026-11-22"
        if fmt == "email":
            return "fighter@example.com"
        if fmt == "uri":
            return "https://example.com"
        title = str(schema.get("title", "")).lower()
        return "id" if title.endswith("id") else "text"
    return None


def group_for(path: str) -> str:
    if path in ("/", "/healthz", "/readyz", "/api/v1/version"):
        return "Service"
    if "/auth/" in path:
        return "Authentication"
    if "/fights/" in path:
        return "Live scoring"
    if "/signup" in path or "/registration" in path:
        return "Registration"
    if "/tournaments" in path or "/templates" in path:
        return "Setup & templates"
    if "/uploads" in path or "/upload" in path:
        return "Media uploads"
    if any(part in path for part in ("/query", "/audit", "/admin", "/permissions", "/export")):
        return "Administration"
    if "/demo/" in path or "/simulate" in path:
        return "Demo & operations"
    return "Events & displays"


def parameter_value(name: str, schema: dict) -> object:
    if name == "eventId":
        return "ev-okami"
    if name in ("fightId", "id"):
        return "fight-01"
    if name.endswith("Id"):
        return "id"
    if name == "table":
        return "fighters"
    example = example_for(schema, {})
    if example is None:
        example = "text" if schema.get("type", "string") == "string" else 0
    return example


def load_operations(spec: dict):
    components = spec.get("components", {})
    default_security = spec.get("security", [])
    operations = []
    for path, methods in spec.get("paths", {}).items():
        for method, operation in methods.items():
            if method.lower() not in METHODS:
                continue
            params = []
            for param in operation.get("parameters", []):
                params.append({
                    "name": param.get("name", ""),
                    "in": param.get("in", "query"),
                    "required": bool(param.get("required", param.get("in") == "path")),
                    "description": param.get("description", ""),
                    "value": parameter_value(param.get("name", ""), param.get("schema", {})),
                })
            body_spec = operation.get("requestBody", {}).get("content", {}).get("application/json", {})
            body = body_spec.get("example")
            if body is None:
                body = example_for(body_spec.get("schema", {}), components)
            operations.append({
                "group": group_for(path),
                "method": method.upper(),
                "path": path,
                "operationId": operation.get("operationId", ""),
                "title": operation.get("summary", operation.get("operationId", path)),
                "description": operation.get("description", ""),
                "auth": bool(operation.get("security", default_security)),
                "parameters": params,
                "body": json.dumps(body, indent=2) if body is not None else "",
            })
    operations.sort(key=lambda op: (op["group"], op["path"], op["method"]))
    return operations


def load_search(manifest: dict, docs_dir: Path, operations: list):
    entries = []
    for doc in manifest["docs"]:
        url = doc.get("path", "/docs/" + doc["slug"])
        entries.append({
            "title": doc["title"],
            "section": doc["group"],
            "text": doc["description"],
            "url": url,
        })
        filename = doc.get("file")
        if not filename:
            continue
        source = (docs_dir / filename).read_text(encoding="utf-8") if (docs_dir / filename).exists() else ""
        current = {"title": doc["title"], "lines": []}
        for line in source.splitlines() + [""]:
            heading = re.match(r"^##\s+(.+)$", line)
            if heading:
                if current["lines"]:
                    text = " ".join(current["lines"]).strip()
                    if text:
                        entries.append({
                            "title": f"{doc['title']} · {current['title']}",
                            "section": doc["group"],
                            "text": text[:400],
                            "url": f"{url}#{slugify(current['title'])}",
                        })
                current = {"title": heading.group(1).replace("`", ""), "lines": []}
            else:
                current["lines"].append(line.strip())
        if current["lines"]:
            text = " ".join(current["lines"]).strip()
            if text:
                entries.append({
                    "title": f"{doc['title']} · {current['title']}",
                    "section": doc["group"],
                    "text": text[:400],
                    "url": f"{url}#{slugify(current['title'])}",
                })
    for op in operations:
        entries.append({
            "title": f"{op['method']} {op['path']}",
            "section": "API reference",
            "text": op["title"],
            "url": f"/docs/api?path={op['path']}&method={op['method']}",
        })
    return entries


def main():
    if len(sys.argv) != 4:
        raise SystemExit("usage: build-docs-index.py <openapi.yaml> <static-docs-dir> <manifest.json>")
    spec_path, docs_dir, manifest_path = map(Path, sys.argv[1:])
    docs_dir.mkdir(parents=True, exist_ok=True)
    spec = yaml.safe_load(spec_path.read_text(encoding="utf-8"))
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    operations = load_operations(spec)
    (docs_dir / "operations.json").write_text(json.dumps(operations, indent=2) + "\n", encoding="utf-8")
    (docs_dir / "search.json").write_text(json.dumps(load_search(manifest, docs_dir, operations), indent=2) + "\n", encoding="utf-8")
    print(f"✅ Generated {len(operations)} API operations and the docs search index")


if __name__ == "__main__":
    main()
