# fastplate
## Description
A fast templating engine for text files.  
No matter if you need to unclutter massive files and split them up, do basic calculations or loop over variables, 
fastplate got you covered!

## Setup
To use fastplate, download the latest [release](https://github.com/xIRoXaSx/fastplate/releases) or clone and build the project locally.  
After that, create templates of your files and give it a run with the required [options](#cli-options).  
Fitting your requirements, you can also use optional [variables](#variables), [functions](#functions) and [loop](#loops) features.  

## CLI options
You can pass the listed arguments / options down below.  

| Argument        | Description                                                                                    |
|-----------------|------------------------------------------------------------------------------------------------|
| -in {FilePath}  | The input path to complete.                                                                    |
| -out {FilePath} | The output path for the completed template(s).                                                 |
| -var {FilePath} | The optional variable file path for unscoped variables.                                        |
| -no-stats       | Disable stats printing.                                                                        |
| -indent         | Enable indention. Spaces / tabs in front of `import` statements will be used for the partials. |
| -crlf           | Split and join contents by CRLF (\r\n) instead of LF (\n).                                     |

### Example
1. Complete template "tempalte.json":  
   `fastplate -in tempalte.json -out completed.json -var fastplate.var`

---

2. Complete multiple templates inside the directory "src":  
   `fastplate -in src/ -out dest/ -indent`

---

## Syntax
The syntax for various interpretations are shown in the table down below.  
The prefix `# fastplate` is always needed for fastplate's  interpretations and can also be used in form of `#fastplate`.  

| Syntax                  | Description                                                                                                        |
|-------------------------|--------------------------------------------------------------------------------------------------------------------|
| import {FilePath}       | Import a file into the current template / partial. Paths are always relational to the working dir.                 |
| var {Name} = {Value}    | Declare a scoped variable of the name `{Name}` and the value `{Value}`.                                            |
| ignore {start / end}    | Starts / ends a ignore block. Lines between these declarations will not be written to the output file.             |
| foreach [{{var1}}, ...] | Loops over each variable until it hits `foreachend`. Use `{{value}}` and `{{index}}` respectively inside the loop. |
| foreachend              | Signals the end of the foreach loop.                                                                               |

### Variables
Import variables can be declared and used from inside the template / partial file (= scoped / local) or 
via an additional variable file (default: `fastplate.var`), which variables can be used throughout every 
template / partial (= unscoped / global).  
The syntax to declare both variable types is the same.  
fastplate automatically looks for `fastplate.var` in the current working directory. If existing, you can use the unscoped 
variables without passing in the `-var` argument.

#### Functions
Functions can be combined / nested like in the following example: `{{func_1(arg1, arg2, {{func_2(arg3, arg4)}})}}`.
You can use the following functions for any type of variable or static values:

| Function name | Description                                                                                   | Example                                |
|---------------|-----------------------------------------------------------------------------------------------|----------------------------------------|
| add()         | Adds the given numbers (variable or static values possible).                                  | `{{add(varName, ...)}}`                |
| sub()         | Subtracts the given numbers from the first one (variable or static values possible).          | `{{sub(varName, ...)}}`                |
| mult()        | Multiplies the given numbers (variable or static values possible).                            | `{{mult(varName, ...)}}`               |
| div()         | Divides the given numbers from the first one (variable or static values possible).            | `{{div(varName, ...)}}`                |
| pow()         | Calculates the power of the given values (first arg = base, second arg = exponent).           | `{{pow(varName, ...)}}`                |
| sqrt()        | Calculates the square root of the given value.                                                | `{{sqrt(varName, ...)}}`               |
| max()         | Chooses the maximum of the given numbers (variable or static values possible).                | `{{max(varName, ...)}}`                |
| min()         | Chooses the minimum of the given numbers (variable or static values possible).                | `{{min(varName, ...)}}`                |
| mod()         | Calculates the modulo (variable or static values possible).                                   | `{{mod(varName, ...)}}`                |
| modmin()      | Same as `mod` but defaults to `min` when remainder is 0 (variable or static values possible). | `{{modmin(varName, ..., min)}}`        |
| floor()       | Rounds down the given value to the nearest integer value.                                     | `{{floor(varName)}}`                   |
| ceil()        | Rounds up the given value to the nearest integer value.                                       | `{{ceil(varName)}}`                    |
| round()       | Rounds the given value to the nearest integer value.                                          | `{{round(varName)}}`                   |
| fixed()       | Rounds the given float value to the given `decimal` place.                                    | `{{fixed(varName, decimal)}}`          |
| sha1()        | Calculates the SHA1 sum of the given file.                                                    | `{{sha1(file_path)}}`                  |
| sha256()      | Calculates the SHA256 sum of the given file.                                                  | `{{sha256(file_path)}}`                |
| sha512()      | Calculates the SHA256 sum of the given file.                                                  | `{{sha512(file_path)}}`                |
| md5()         | Calculates the MD5 sum of the given file.                                                     | `{{md5(file_path)}}`                   |
| now()         | Prints the time of execution in the given [format](https://pkg.go.dev/time#pkg-constants).    | `{{now(format)}}`                      |
| lower()       | Prints the variable's value in lower case.                                                    | `{{lower(varName)}}`                   |
| upper()       | Prints the variable's value in upper case.                                                    | `{{upper(varName)}}`                   |
| cap()         | Prints the first letter of each word of the variable's value in upper case.                   | `{{cap(varName)}}`                     |
| split()       | Splits the value by `seperator` and print the element at `index`.                             | `{{split(varName, seperator, index)}}` |
| repeat()      | Repeats the given value `amount` times.                                                       | `{{repeat(varName, amount)}}`          |
| replace()     | Replaces `old` in the given value `value` with `new`.                                         | `{{replace(value, old, new)}}`         |
| len()         | Either prints the length or the amount of variables (`UNSCOPED_VARS`) of the given value      | `{{len(varName)}}`                     |
| var()         | Creates a new scoped variable which can be used after the declaration.                        | `{{var(varName, value)}}`              |

### Loops
Looping over multiple variables can be implemented by using the `foreach` syntax.  
For every iteration you can retrieve the index with `{{index}}` and the value with `{{value}}`.  
By declaring variables (no matter if scoped or unscoped), you can loop over selected ones like so (`[]` brackets are optional):  
```
# fastplate foreach [ {{var1}}, {{var2}}, {{var3}}, {{unscoped_threshold}} ]
   Insert your value to repeat here.
# fastplate foreachend 
```

If you have countless variables, you can put all of those variables into a dedicated [var files](#variables) and use 
the special variable `{{UNSCOPED_VARS}}`.  
This way, fastplate will loop over each unscoped variable automatically (`[]` brackets are optional):
```
# fastplate foreach [ {{UNSCOPED_VARS}} ]
   Insert your value to repeat here.
# fastplate foreachend 
```

In addition to the latter option, you can also restrict the variables to use to one specific variable file.  
In order to do so, you need to add the prefix `_` and the base name of your var file (without the extension, case-insensitive) 
like in this example (`[]` brackets are optional):  
```
# fastplate foreach [ {{UNSCOPED_VARS_yourVarFileName}} ]
   Insert your value to repeat here.
# fastplate foreachend 
```

These special variables are currently only supported for the `foreach` loop!

You can also use an integer value for the foreach loop to use it as a for 0 - n loop.  
The value needs to be either statically typed (`5`) or stored in a variable (`{{iterations}}`).  
Here is an example (`[]` brackets are optional):  
```
# fastplate foreach [ 5 ]
   Insert your value to repeat here.
# fastplate foreachend

OR

# fastplate iterations = 5
# fastplate foreach [ {{iterations}} ]
   Insert your value to repeat here.
# fastplate foreachend
```

TIP:
For nested loops, you can also use the `var` function to create a dynamic scoped variable.  
This way you are able to use the outer `index`, `value` and `name` variables inside a child loop.  
Here is an example (`[]` brackets are optional):
```
# fastplate foreach [ {{var1}}, {{var2}}, {{var3}}, {{unscoped_threshold}} ]
   {{var(outerIndex, index)}}
   # fastplate foreach [ {{var1}}, {{var2}}, {{var3}}, {{unscoped_threshold}} ]
      Outer index: {{outerIndex}}, inner index: {{index}}
   # fastplate foreachend 
# fastplate foreachend 
```

### Example
1. Import `src/partials/world.txt` (which contains "World!") into the current template.
```text
Hello
   # fastplate import src/partials/world.txt
```

Result:
```text
Hello
   World!
```

---

2. Declare and use a scoped variable:
```text
# fastplate var world = World!
Hello
   {{world}}
```

Result:
```text
Hello
   World!
```

---

3. Declare and use an unscoped variable:  
   File fastplate.var:
```text
# fastplate var world = World!
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

4. Use a foreach loop to iterate over specific variables (`[]` brackets are optional), may also be used with unscoped vars:
```text
Shopping list:
# fastplate var apples = Apples
# fastplate var oranges = Oranges
# fastplate var bananas = Bananas
# fastplate foreach [ {{apples}}, {{oranges}}, {{bananas}} ]
  {{index}}.) 2x {{value}}
# fastplate foreachend
```

Result:
```text
Shopping list:
  0.) 2x Apples
  1.) 2x Oranges
  2.) 2x Bananas
```

---

5. Use a foreach loop to iterate over every **unscoped** variable (`[]` brackets are optional):  
   File fastplate.var:
```text
# fastplate var hello = Hello
# fastplate var world = World!
```

```text
# fastplate foreach [ {{UNSCOPED_VARS}} ]
  {{index}} -> {{value}}
# fastplate foreachend
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
  # fastplate var apples = APPLES
  # fastplate var oranges = oranges
  # fastplate var bananas = bananas
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
