import os, glob, re
paths = glob.glob('pkg/game/quest/*_test.go') + glob.glob('internal/api/quests/*_test.go')
for p in paths:
    with open(p, 'r') as f:
        content = f.read()
    content = re.sub(r'h\.CompleteChallenge\(([^,]+),\s*([^,]+),\s*([^,]+),\s*([^,]+),\s*([^,\)]+)\)', r'h.CompleteChallenge(\1, \2, \3, \4, \5, "")', content)
    content = re.sub(r'svc\.CompleteChallengeForQuest\(([^,]+),\s*([^,]+),\s*([^,]+),\s*([^,\)]+)\)', r'svc.CompleteChallengeForQuest(\1, \2, \3, \4, "")', content)
    content = re.sub(r'handler\.CompleteChallenge\(([^,]+),\s*([^,]+),\s*([^,]+),\s*([^,]+),\s*([^,\)]+)\)', r'handler.CompleteChallenge(\1, \2, \3, \4, \5, "")', content)
    with open(p, 'w') as f:
        f.write(content)
