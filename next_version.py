#!/usr/bin/env python3
"""
    Script to automate the process of versioning in a git repository.
    It calculates the next version based on the latest git tag and creates a new
    tag.
"""

import subprocess
import sys


def run_command(cmd):
    """Run a shell command and return the output."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout.strip(), result.returncode


def get_latest_tag():
    """Get the latest git tag."""
    output, _ = run_command("git tag --sort=-v:refname | head -1")
    return output if output else None


def calculate_next_version(latest_tag):
    """Calculate the next patch version."""
    if not latest_tag:
        return "v0.1.0"

    # Remove 'v' prefix if present
    version = latest_tag.lstrip('v')

    # Split version into major.minor.patch
    parts = version.split('.')
    if len(parts) != 3:
        print(f"Error: Invalid version format: {latest_tag}")
        sys.exit(1)

    major, minor, patch = parts

    # Increment patch version
    next_patch = int(patch) + 1
    return f"v{major}.{minor}.{next_patch}"


def create_and_push_tag(tag):
    """Create and push the git tag."""
    # Create the tag
    _, returncode = run_command(f"git tag {tag}")
    if returncode != 0:
        print("Failed to create tag")
        return False

    print(f"Tag {tag} created successfully")

    # Push the tag to remote
    _, returncode = run_command(f"git push origin {tag}")
    if returncode != 0:
        print("Failed to push tag to remote")
        return False

    print(f"Tag {tag} pushed to remote successfully")
    return True


def main():
    # Get the latest tag
    latest_tag = get_latest_tag()
    print(f"Latest tag: {latest_tag if latest_tag else 'none'}")

    # Calculate next version
    next_version = calculate_next_version(latest_tag)
    print(f"Next version: {next_version}")

    # Ask for confirmation
    response = input(f"Do you want to create and push tag {next_version}? (y/n): ")

    if response.lower() in ['y', 'yes']:
        if create_and_push_tag(next_version):
            sys.exit(0)
        else:
            sys.exit(1)
    else:
        print("Tag creation cancelled")


if __name__ == "__main__":
    main()
