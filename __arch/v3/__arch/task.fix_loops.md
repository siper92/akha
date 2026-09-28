# task

1. In cmd/utils - write a commands that read the given path and makes a module source map 
 - see: _env/dev_cache/*.goyml file
 - command in utils are different module so have their own main
   - for now the only command is: akha_utils source-map --path <path> --output <output_path> --with-private
   - parse recursively the given path
   - with-private - include private funcs and types in the source map
   - for a module produce a source_map_<module_name>_<time>.goyml file in the output path
 - add a just file command to run the command for the - cmd, internal, lang and platform module
   - the command must firs clear the output path and then run the command - use rm -rf in just 
   - the command overrides files if present