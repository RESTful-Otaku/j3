import yaml, pathlib, sys
root=pathlib.Path('.github/workflows')
errors=[]
for p in root.glob('*.yml'):
    try:
        data=yaml.safe_load(p.read_text())
    except Exception as e:
        errors.append(f'{p}: invalid YAML {e}')
        continue
    for job_name, job in (data.get('jobs') or {}).items():
        uses=(job or {}).get('uses','')
        if uses.startswith('./.github/workflows/'):
            target=pathlib.Path(uses[2:])
            called=yaml.safe_load(target.read_text()) if target.exists() else None
            if not called or 'workflow_call' not in (called.get('on') or {} if isinstance(called.get('on'),dict) else {}):
                # PyYAML may parse 'on' as True in YAML 1.1; handle both
                if not called or (called.get(True) is None and 'workflow_call' not in called.get('on',{})):
                    errors.append(f'{p}:{job_name} calls {target} without workflow_call')
print('\n'.join(errors) if errors else 'workflows ok')
sys.exit(1 if errors else 0)
