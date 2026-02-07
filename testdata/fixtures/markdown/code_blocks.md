# Code Examples

## Go

```go
package main

import "fmt"

func main() {
	for i := 0; i < 10; i++ {
		fmt.Printf("Hello %d\n", i)
	}
}
```

## Python

```python
def fibonacci(n: int) -> list[int]:
    """Generate fibonacci sequence."""
    a, b = 0, 1
    result = []
    for _ in range(n):
        result.append(a)
        a, b = b, a + b
    return result
```

## JavaScript

```javascript
async function fetchData(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }
  return response.json();
}
```

## Shell

```bash
#!/bin/bash
set -euo pipefail

for file in *.md; do
    echo "Processing: $file"
    wc -l "$file"
done
```

## SQL

```sql
SELECT u.name, COUNT(o.id) AS order_count
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
WHERE u.created_at > '2025-01-01'
GROUP BY u.name
HAVING COUNT(o.id) > 5
ORDER BY order_count DESC;
```

## Inline Code

Use `fmt.Println()` in Go or `print()` in Python.
