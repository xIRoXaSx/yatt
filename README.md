# yatt
## Description
Yet another templating engine for text files.  
No matter if you need to unclutter massive files and split them up, do basic calculations or loop over variables, yatt got you covered!

## Setup
To use yatt, download the latest [release](https://github.com/xIRoXaSx/yatt/releases) or clone 
and build (`CGO_ENABLED=0 go build -o="yatt" .`) the project locally.  
After that, create templates of your files and give it a run with either a start file or directory.  
Fitting your requirements, you can use optional [variables](#variables), [functions](#functions) or [preprocessors](#preprocessors).  

## CLI options
You can pass the listed arguments / options down below.  

| Argument        | Description                                                                                    |
|-----------------|------------------------------------------------------------------------------------------------|
| -in {FilePath}  | The input path of your template(s) to complete.                                                |
| -out {FilePath} | The output path for the completed template(s).                                                 |
| -var {FilePath} | The optional variable file path for global variables.                                          |
| -blacklist      | Regex pattern(s) to describe which files should not be interpreted.                            |
| -whitelist      | Regex pattern(s) to describe which files should be interpreted .                               |
| -verbose        | Enables the verbose print option.                                                              |
| -no-stats       | Disable stats printing.                                                                        |
| -indent         | Enable indention. Spaces / tabs in front of `import` statements will be used for the partials. |
| -crlf           | Split and join contents by CRLF (\r\n) instead of LF (\n).                                     |

### Usage
1. Complete template "template.json":  
   `yatt -in template.json -out completed.json -var yatt.var`

---

2. Complete multiple templates inside the directory "src":  
   `yatt -in src/ -out dest/ -indent`

---

## Syntax
### Preprocessors
Preprocessors can be used to manipulate text before it gets interpreted.  
The prefix `# yatt` or `// yatt` is always required for interpretations.  
The following table contains all available operations:  

| Preprocessor         | Description                                                                                        | Example                                  |
|----------------------|----------------------------------------------------------------------------------------------------|------------------------------------------|
| import               | Import a file into the current template / partial. Paths are always relational to the working dir. | `# yatt import my/test/file.txt`         |
| var                  | Declare a scoped variable of the name `{Name}` and the value `{Value}`.                            | `# yatt var myVar = 123`                 |
| ignore / ignoreend   | Starts / ends a ignore block. Lines between these declarations will not be written to the output.  | `# yatt ignore` ... `# yatt ignoreend`   |
| foreach / foreachend | Loops over each variable until `foreachend`. Use `{{value}}` and `{{index}}` inside the loop.      | `# yatt foreach` ... `# yatt foreachend` |

### Variables
Variables can be declared and used from inside the templated file (local, can only be used inside this file) or via an additional file, 
which variables can be used throughout every template (global variables).  
The syntax for both variable scopes is identical.  

#### Functions
Functions can be combined / nested as you like, e.g.: `{{func_1(arg1, arg2, {{func_2(arg3, arg4)}})}}`.
You can use the following functions for any type of variable or static values:

| Function name | Description                                                                                     | Example                                |
|---------------|-------------------------------------------------------------------------------------------------|----------------------------------------|
| add()         | Adds the given numbers (variable or static values possible).                                    | `{{add(varName, ...)}}`                |
| sub()         | Subtracts the given numbers from the first one (variable or static values possible).            | `{{sub(varName, ...)}}`                |
| mult()        | Multiplies the given numbers (variable or static values possible).                              | `{{mult(varName, ...)}}`               |
| div()         | Divides the given numbers from the first one (variable or static values possible).              | `{{div(varName, ...)}}`                |
| pow()         | Calculates the power of the given values (first arg = base, second arg = exponent).             | `{{pow(varName, ...)}}`                |
| sqrt()        | Calculates the square root of the given value.                                                  | `{{sqrt(varName, ...)}}`               |
| max()         | Chooses the maximum of the given numbers (variable or static values possible).                  | `{{max(varName, ...)}}`                |
| min()         | Chooses the minimum of the given numbers (variable or static values possible).                  | `{{min(varName, ...)}}`                |
| mod()         | Calculates the modulo (variable or static values possible).                                     | `{{mod(varName, ...)}}`                |
| env()         | Prints the value of the given environment variable.                                             | `{{env(ENV_VAR)}}`                     |
| floor()       | Rounds down the given value to the nearest integer value.                                       | `{{floor(varName)}}`                   |
| ceil()        | Rounds up the given value to the nearest integer value.                                         | `{{ceil(varName)}}`                    |
| round()       | Rounds the given value to the nearest integer value.                                            | `{{round(varName)}}`                   |
| fixed()       | Rounds the given float value to the given `decimal` place.                                      | `{{fixed(varName, decimal)}}`          |
| sha1()        | Calculates the SHA1 sum of the given file.                                                      | `{{sha1(file_path)}}`                  |
| sha256()      | Calculates the SHA256 sum of the given file.                                                    | `{{sha256(file_path)}}`                |
| sha512()      | Calculates the SHA256 sum of the given file.                                                    | `{{sha512(file_path)}}`                |
| md5()         | Calculates the MD5 sum of the given file.                                                       | `{{md5(file_path)}}`                   |
| now()         | Prints the time of execution in the given [format](https://github.com/xIRoXaSx/godate#formats). | `{{now(format)}}`                      |
| lower()       | Prints the variable's value in lower case.                                                      | `{{lower(varName)}}`                   |
| upper()       | Prints the variable's value in upper case.                                                      | `{{upper(varName)}}`                   |
| cap()         | Prints the first letter of each word of the variable's value in upper case.                     | `{{cap(varName)}}`                     |
| split()       | Splits the value by `seperator` and print the element at `index`.                               | `{{split(varName, seperator, index)}}` |
| repeat()      | Repeats the given value `amount` times.                                                         | `{{repeat(varName, amount)}}`          |
| replace()     | Replaces `old` in the given value `value` with `new`.                                           | `{{replace(value, old, new)}}`         |
| basename()    | Prints the current file's base name (filename + extension).                                     | `{{basename()}}`                       |
| name()        | Prints the current file's name (relative path included).                                        | `{{name()}}`                           |
| len()         | Either prints the length of the given value or the amount of variables (`YATT_VARS`).           | `{{len(varName)}}`                     |
| var()         | Creates a new local variable which can be used after the declaration.                           | `{{var(varName, value)}}`              |

### Loops
Looping over multiple variables can be done by using the `foreach` syntax.  
For every iteration of a foreach loop, you can retrieve the index with `{{index}}` and the value with `{{value}}`.  
By declaring variables, you can loop over selected ones like so (`[]` brackets are optional):  
```
# yatt foreach [ {{var1}}, {{var2}}, {{var3}}, {{global_threshold}} ]
   Insert your value to repeat here.
# yatt foreachend
```

If you have countless variables, you can put all of those variables into a dedicated file (described in [variables](#variables)) and use 
the special variable `{{YATT_VARS}}`.  
This way, yatt will loop over each global variable automatically (`[]` brackets are optional):
```
# yatt foreach [ {{YATT_VARS}} ]
   Insert your value to repeat here.
# yatt foreachend 
```

In addition to the latter option, you can also restrict the loop to use variables of one specific file.  
In order to do so, you need to add `_` and the file path of your file like in this example (`[]` brackets are optional):  
```
# yatt foreach [ {{YATT_VARS_myVariables.txt}} ]
   Insert your value to repeat here.
# yatt foreachend 
```

These special variables are currently only supported for the `foreach` loop!

You can also use an integer value for the foreach loop to use it as a for 0 - n loop.  
The value needs to be either statically typed (`5`) or stored in a variable (e.g.: `{{iterations}}`).  
For every iteration of a foreach loop, only the `{{index}}` variable is dynamically created.  
Here is an example (`[]` brackets are optional):  
```
# yatt foreach [ 5 ]
   Insert your value to repeat here.
# yatt foreachend

OR

# yatt iterations = 5
# yatt foreach [ {{iterations}} ]
   Insert your value to repeat here.
# yatt foreachend
```

TIP:
For nested loops, you can also use the `var` function to create a dynamic, foreach-scoped variable.  
This way you are able to use the outer `index` and `value` variables inside a child loop.  
Here is an example (`[]` brackets are optional):
```
# yatt foreach [ {{var1}}, {{var2}}, {{var3}}, {{global_threshold}} ]
   {{var(outerIndex, index)}}
   # yatt foreach [ {{var1}}, {{var2}}, {{var3}}, {{global_threshold}} ]
      Outer index: {{outerIndex}}, inner index: {{index}}
   # yatt foreachend 
# yatt foreachend 
```

### Examples
1. Import `src/partials/world.txt` (which contains "World!") into the current template.
```text
Hello
   # yatt import src/partials/world.txt
```

Result:
```text
Hello
   World!
```

---

2. Declare and use a local variable:
```text
# yatt var world = World!
Hello
   {{world}}
```

Result:
```text
Hello
   World!
```

---

3. Declare and use an global variable:  
   File yatt.var:
```text
# yatt var world = World!
```

Template:
```text
Hello
   {{world}}
```
Result:
```text
Hello
   World!
```

---

4. Use a foreach loop to iterate over specific variables (`[]` brackets are optional):
```text
Shopping list:
# yatt var apples = Apples
# yatt var oranges = Oranges
# yatt var bananas = Bananas
# yatt foreach [ {{apples}}, {{oranges}}, {{bananas}} ]
  {{index}}.) 2x {{value}}
# yatt foreachend
```

Result:
```text
Shopping list:
  0.) 2x Apples
  1.) 2x Oranges
  2.) 2x Bananas
```

---

5. Use a foreach loop to iterate over **every global** variable (`[]` brackets are optional):  
   File yatt.var:
```text
# yatt var hello = Hello
# yatt var world = World!
```

```text
# yatt foreach [ {{YATT_VARS}} ]
  {{index}} -> {{value}}
# yatt foreachend
```

Result:
```text
  0 -> Hello
  1 -> World!
```

---

6. Use functions:
```text
Shopping list:
  # yatt var apples = APPLES
  # yatt var oranges = oranges
  # yatt var bananas = bananas
  2x {{lower(apples)}}
  2x {{upper(oranges)}}
  2x {{cap(bananas)}}
```

Result:
```text
Shopping list:
  2x apples
  2x ORANGES
  2x Bananas
```
