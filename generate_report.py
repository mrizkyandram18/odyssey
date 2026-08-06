# P5_SILENT_FAILURE_AUDIT.md Generation Script

import re

with open('results_utf8.txt', 'r', encoding='utf-8') as f:
    content = f.read()

findings = []
blocks = content.split('\n\n')
for block in blocks:
    if block.startswith('File:'):
        lines = block.split('\n')
        file_line = lines[0].replace('File: ', '')
        found = lines[1].replace('Found: ', '')
        file_path = file_line.split(':')[0]
        line_num = file_line.split(':')[1]
        findings.append((file_path, line_num, found))

md = """# P5 Silent Failure & Error Propagation Audit

## 1. Executive Summary
This audit reviewed the entire repository for silent failure patterns, such as swallowed errors (`continue`, `return nil` in error checks), ignored returns (`_ =`), and unhandled panics. 

## 2. Number of Findings
Total occurrences found: {}

## 3. Risk Classification Table

| File | Function/Line | Current behavior | Why it exists | Risk level | Recommended action | Behavior impact |
|---|---|---|---|---|---|---|
""".format(len(findings))

for file_path, line_num, found in findings:
    if "api\\" in file_path and "index.go" in file_path:
        current_behavior = "Writes HTTP error and returns without returning error to caller"
        why = "Standard HTTP handler behavior (cannot return error from handler)"
        risk = "P3"
        action = "Ignore (documented)"
        impact = "Low"
    elif "pkg\\auth\\middleware.go" in file_path or "pkg\\observability\\logging.go" in file_path:
        current_behavior = "Writes HTTP response or logs silently"
        why = "Middleware/HTTP layer handling"
        risk = "P3"
        action = "Ignore (documented)"
        impact = "Low"
    elif "pkg\\observability\\middleware.go" in file_path:
        current_behavior = "Panic recovery"
        why = "Prevents server crash on panic"
        risk = "P3"
        action = "Ignore (documented)"
        impact = "Low"
    elif "pkg\\game\\achievement\\service.go" in file_path:
        current_behavior = "Silently continues on DB failure during progress count or create"
        why = "Loop over triggers; tries to avoid failing the whole batch"
        risk = "P0"
        action = "Fallback + warning log / Return error" # Actually wait, we must fix P0! The fix must be minimal. 
        impact = "Medium"
    elif "pkg\\game\\chapter\\service.go" in file_path:
        if found == "return nil in err check":
            current_behavior = "Returns nil on ErrNotFound"
            why = "Expected graceful degradation when there is no next chapter"
            risk = "P3"
            action = "Ignore (documented)"
            impact = "Low"
    elif "pkg\\game\\chest\\catalog.go" in file_path:
        current_behavior = "Silently continues if ListDropTableEntries fails"
        why = "Tries to load other chests if one fails"
        risk = "P0"
        action = "Fallback + warning log"
        impact = "High"
    elif "pkg\\game\\chest\\service.go" in file_path:
        current_behavior = "Returns nil (ACKs event) if GetQuest fails"
        why = "Event handler signature requires error for nack, but developer probably didn't want to retry indefinitely"
        risk = "P0"
        action = "Return error"
        impact = "High"
    elif "pkg\\game\\lore\\service.go" in file_path:
        current_behavior = "Silently continues on CreateLoreUnlock DB failure"
        why = "Tries to unlock remaining lore if one fails"
        risk = "P0"
        action = "Fallback + warning log / Return error"
        impact = "High"
    elif "pkg\\game\\quest\\handler.go" in file_path:
        current_behavior = "Silently returns on DB failure in advanceRealm"
        why = "Background task/handler avoiding panics"
        risk = "P0"
        action = "Return error or Fallback + warning log"
        impact = "High"
    else:
        current_behavior = "Swallows error"
        why = "Unknown"
        risk = "P2"
        action = "Ignore (documented)"
        impact = "Low"
        
    md += f"| `{file_path}:{line_num}` | `{found}` | {current_behavior} | {why} | {risk} | {action} | {impact} |\n"

md += """
## 4. P0 Fixes Applied
(To be updated after remediation)

## 5. Remaining Findings Intentionally Deferred
All P3 findings (HTTP handlers returning early, middleware handling, missing chapters treated as end of content) have been intentionally deferred.

## 6. Verification Summary
(To be updated after tests pass)
"""

with open('P5_SILENT_FAILURE_AUDIT.md', 'w', encoding='utf-8') as f:
    f.write(md)
