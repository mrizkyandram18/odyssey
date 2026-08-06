import os

def audit_directory(dir_path):
    issues = []
    
    # Regex to find 'if err != nil' followed by block containing continue, break, return nil
    for root, _, files in os.walk(dir_path):
        for file in files:
            if not file.endswith('.go') or file.endswith('_test.go'):
                continue
                
            filepath = os.path.join(root, file)
            with open(filepath, 'r', encoding='utf-8') as f:
                content = f.read()
                
            # split by lines
            lines = content.split('\n')
            
            in_err_block = False
            block_start = 0
            brace_count = 0
            
            for i, line in enumerate(lines):
                line_stripped = line.strip()
                
                if not in_err_block:
                    if 'if err != nil' in line:
                        in_err_block = True
                        block_start = i
                        brace_count = line.count('{') - line.count('}')
                else:
                    brace_count += line.count('{') - line.count('}')
                    
                    if 'continue' in line_stripped and not line_stripped.startswith('//'):
                        issues.append((filepath, i+1, 'continue in err!=nil block', line_stripped))
                    elif 'return nil' in line_stripped and not line_stripped.startswith('//'):
                        issues.append((filepath, i+1, 'return nil in err!=nil block', line_stripped))
                    elif 'break' in line_stripped and not line_stripped.startswith('//'):
                        issues.append((filepath, i+1, 'break in err!=nil block', line_stripped))
                        
                    if brace_count <= 0:
                        in_err_block = False

    return issues

issues = audit_directory('pkg/game')
for issue in issues:
    print(f"{issue[0]}:{issue[1]} -> {issue[2]}: {issue[3]}")
