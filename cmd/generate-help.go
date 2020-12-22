package cmd

const genHelp = `
Generate Code from folder or Git repository

The template folder should contain a generator.yaml file
with the following fields:

project: bool (optional, default false)
  Determines whether this template is a project or not.
  A project template will check that the destination folder doesn't exists,
  everything else is the same.

data: User Data that is a map of keys as strings and values as strings
  You can store any values to use in the template.
  These values can be overriden by the final user.

files: Array of strings
  List of files that are templates to process.
  For performance considerations, instead of processing
  all files.

actions: Array of actions

Action Type
  Is an object type with the following fields:
  - type: string
  - args: Array of strings

Supported Actions:
  - type: delete
    args:
      - FILE_OR_DIRECTORY
  - type: rename
    args:
      - SOURCE_FILE_OR_DIRECTORY
      - TARGET_TEMPLATE
  - type: copy
    args:
      - SOURCE_FILE
      - TARGET_TEMPLATE
  - type: insert-after
    args:
      - SOURCE_FILE
      - SEARCH_REGEXP
      - LINE_TEMPLATE
  - type: insert-before
    args:
      - SOURCE_FILE
      - SEARCH_REGEXP
      - LINE_TEMPLATE
  - type: replace-all
    args:
      - SOURCE_FILE
      - SEARCH_REGEXP
      - REPLACE_TEMPLATE

Template fields uses Go Text Template Idiom with the user data as a context

Go Templates reference
	- https://curtisvermeeren.github.io/2017/09/14/Golang-Templates-Cheatsheet
	- https://blog.gopheracademy.com/advent-2017/using-go-templates/

`
