# Todowner

## Overview

Todowner is a tool that processes `.todo` files in a given folder recursively, creates backups of these files, and converts them to markdown format.

## Features

- Recursively find `.todo` files in a given folder.
- Create a backup of the `.todo` files.
- Convert `.todo` files to markdown format.

## Usage

### Main Function

The main function performs the following tasks:

1. **Find the `.todo` files in a given folder, recursively.**
    - Skips hidden directories.
    - Collects the paths of all `.todo` files.

2. **Create the backup folder.**
    - Creates a folder named `todowner_backup` to store the backups of the `.todo` files.

3. **Process each `.todo` file.**
    - Creates the full nested filepath inside the backup folder.
    - Converts todo sections to markdown headings.
    - Converts todo boxes to markdown checkboxes.
    - Adds a warning to lines that are not headings and do not start with a dash.
    - Non-heading lines should have their nesting levels reduced by the number of levels the containing heading had in the 'todo' format.
    - The old format was based on the premise that:
        - Headings end with ':'.
        - Content starts one indentation level after the heading.
        - In markdown, this is not needed, so we also remove one more indentation level in non-heading lines.
