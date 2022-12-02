# moss_plus

moss_plus is a code similarity tester based on similarity-tester and it can generate beautiful html page for better look.

### Usage

Typical usage: `moss_plus <input_dir>/<input_files> --output <output_dir>`

Example usage: `moss_plus <input_dir>/*.* --output <output_dir>`

For C language: `moss_plus <input_dir>/*.c --language c --output <output_dir>`

For C++ languages: `moss_plus <input_dir>/*.cpp --language c++ --output <output_dir>`

For Java languages: `moss_plus <input_dir>/*.java --language java --output <output_dir>`

### Features

1. No test code from same submitter
2. Auto-detect code language and test
3. Output good-looking html page