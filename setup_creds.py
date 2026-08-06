import io
import re

with io.open(r'D:\Personal\Projects\kuota - Copy\.env', 'r', encoding='utf-8') as f:
    kuota = f.read()

fb_cred = re.search(r'FIREBASE_CREDENTIALS="(.*?)"\n', kuota, re.DOTALL).group(1)

# fix the newline literal issues if any
fb_cred = fb_cred.replace('\\n', '\n')

with io.open('creds.json', 'w', encoding='utf-8') as f:
    f.write(fb_cred)

with io.open('.env', 'r', encoding='utf-8') as f:
    env = f.read()

env = re.sub(r'FIREBASE_CREDENTIALS=.*', '', env)
env = re.sub(r'GOOGLE_APPLICATION_CREDENTIALS_JSON=.*', 'GOOGLE_APPLICATION_CREDENTIALS=creds.json', env)

with io.open('.env', 'w', encoding='utf-8') as f:
    f.write(env)
