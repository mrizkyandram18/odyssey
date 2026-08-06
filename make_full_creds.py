import json

# Read fake key
with open('fake_key.txt', 'r', encoding='utf-16le') as f:
    key_pem = f.read().strip()

key_pem = key_pem.replace('\\n', '\n')

creds = {
  "type": "service_account",
  "project_id": "dummy",
  "private_key_id": "dummy_id",
  "private_key": key_pem,
  "client_email": "dummy@dummy.iam.gserviceaccount.com",
  "client_id": "12345",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/dummy%40dummy.iam.gserviceaccount.com",
  "universe_domain": "googleapis.com"
}

with open('creds.json', 'w') as f:
    json.dump(creds, f)

import re
with open('.env', 'r', encoding='utf-8') as f:
    env = f.read()

env = re.sub(r'FIREBASE_CREDENTIALS=.*', '', env)
env = re.sub(r'GOOGLE_APPLICATION_CREDENTIALS_JSON=.*', 'GOOGLE_APPLICATION_CREDENTIALS=creds.json', env)

with open('.env', 'w', encoding='utf-8') as f:
    f.write(env)

print("Generated full creds.json and updated .env")
