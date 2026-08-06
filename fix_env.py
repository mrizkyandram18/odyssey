import io
import json

with io.open('creds.json', 'r', encoding='utf-8') as f:
    creds_str = f.read()

# make sure it's valid json
try:
    json.loads(creds_str)
except:
    print("invalid json in creds.json")

with io.open('.env', 'r', encoding='utf-8') as f:
    env = f.read()

import re
env = re.sub(r'FIREBASE_CREDENTIALS=.*', '', env)
env = re.sub(r'GOOGLE_APPLICATION_CREDENTIALS=.*', '', env)

# just append it
# need to escape newlines inside the json string if we want it on one line, but it's already one line in kuota
# actually creds_str might have real newlines? No, it has \n literals.
env = env.strip() + "\n" + f"FIREBASE_CREDENTIALS='{creds_str}'\n"

with io.open('.env', 'w', encoding='utf-8') as f:
    f.write(env)
