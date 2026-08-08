const { execSync } = require('child_process');
const fs = require('fs');

const envFile = fs.readFileSync('.env', 'utf8');
const lines = envFile.split('\n');

for (const line of lines) {
    const match = line.match(/^([^#]+?)=(.*)$/);
    if (match) {
        const key = match[1].trim();
        let value = match[2].trim();
        if (value.startsWith('"') && value.endsWith('"')) {
            value = value.substring(1, value.length - 1);
        }
        if (key && value) {
            console.log(`Adding ${key}...`);
            try {
                // First try to remove it
                execSync(`npx vercel env rm ${key} production -y`, { stdio: 'ignore' });
            } catch (e) {}
            try {
                // Add it
                execSync(`npx vercel env add ${key} production`, {
                    input: value,
                    stdio: ['pipe', 'inherit', 'inherit']
                });
            } catch (e) {
                console.error(`Failed to add ${key}`);
            }
        }
    }
}
