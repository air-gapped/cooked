"""Sample Python module for testing syntax highlighting."""

from dataclasses import dataclass, field
from typing import Iterator


@dataclass
class TreeNode:
    """A node in a binary tree."""

    value: int
    left: "TreeNode | None" = None
    right: "TreeNode | None" = None
    tags: list[str] = field(default_factory=list)

    def depth(self) -> int:
        """Return the depth of this subtree."""
        left_d = self.left.depth() if self.left else 0
        right_d = self.right.depth() if self.right else 0
        return 1 + max(left_d, right_d)

    def inorder(self) -> Iterator[int]:
        """Yield values in-order."""
        if self.left:
            yield from self.left.inorder()
        yield self.value
        if self.right:
            yield from self.right.inorder()


def build_tree(values: list[int]) -> TreeNode | None:
    """Build a balanced BST from a sorted list."""
    if not values:
        return None
    mid = len(values) // 2
    return TreeNode(
        value=values[mid],
        left=build_tree(values[:mid]),
        right=build_tree(values[mid + 1 :]),
    )


if __name__ == "__main__":
    tree = build_tree([1, 2, 3, 4, 5, 6, 7])
    assert tree is not None
    print(f"Depth: {tree.depth()}")
    print(f"In-order: {list(tree.inorder())}")
