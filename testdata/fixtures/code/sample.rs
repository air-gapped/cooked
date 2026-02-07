//! Sample Rust module for testing syntax highlighting.

use std::collections::HashMap;
use std::fmt;

/// A key-value store with expiration.
#[derive(Debug)]
pub struct Store<V: Clone> {
    data: HashMap<String, Entry<V>>,
}

#[derive(Debug, Clone)]
struct Entry<V: Clone> {
    value: V,
    version: u64,
}

impl<V: Clone + fmt::Display> Store<V> {
    /// Creates a new empty store.
    pub fn new() -> Self {
        Store {
            data: HashMap::new(),
        }
    }

    /// Inserts a value, returning the previous version number.
    pub fn insert(&mut self, key: impl Into<String>, value: V) -> u64 {
        let key = key.into();
        let version = self
            .data
            .get(&key)
            .map(|e| e.version + 1)
            .unwrap_or(1);

        self.data.insert(
            key,
            Entry { value, version },
        );
        version
    }

    /// Gets a reference to a value.
    pub fn get(&self, key: &str) -> Option<&V> {
        self.data.get(key).map(|e| &e.value)
    }

    /// Returns the number of entries.
    pub fn len(&self) -> usize {
        self.data.len()
    }

    /// Returns true if empty.
    pub fn is_empty(&self) -> bool {
        self.data.is_empty()
    }
}

impl<V: Clone + fmt::Display> fmt::Display for Store<V> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for (key, entry) in &self.data {
            writeln!(f, "{}: {} (v{})", key, entry.value, entry.version)?;
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_insert_and_get() {
        let mut store = Store::new();
        store.insert("key", "value");
        assert_eq!(store.get("key"), Some(&"value"));
        assert_eq!(store.len(), 1);
    }
}
