This is a package to do resource planning for an . It takes a directory of model files, and one or more driver values specified on the command line.

The output is, for each model, the computed CPU, RAM, disk bytes, disk IOPS, and number of instances that the model computes.

There must be one model (the "top level" model) named identically to the sub-directory. This is the model that receives the input.

Conceptually, a model has one or more inputs, and zero or more backends.
It also has zero or more resources, computed from the input values. Any input not computed is taken as having the value 0. Each model also has a number of instances (if not specified, this defaults to 0). The expression language understands numbers, addition, subtraction, multiplication and division, in addition to parenthesised sub-expressions and references.

Each model has zero or more inputs, zero or more "outputs" (really, other models whose inputs they feed), zero or more variables, and a resource block, where RAM, CPU and replica count is recorded. The model will not be evaluated until its inputs have been fully populated (thus, no circular dependencies are supported), and it will be evaluated in the order "variables", "outputs", then finally "resources". Variables are evaluated in arbritary order. In an expression, a simple name refers to a variable or input (with inputs having priority in case of conflicts), and the value of a variable is cached (and evaluated "on the spot"). It is thus not a problem having the expression for a variable referencing another variable, but circular references will cause problems.

If no value for a resource (CPU, RAM, replica count) is specified, they will default to 0 for RAM and CPU, and to 1 for the replica count.

A model file is a list of models, one of which should have the same name as the file (without the ".yaml" prefix), this is the 'top-level' object whose inputs are set from the command line.

Each model is on the form:
  name: <name of model>
  inputs:
   - <input1>
   ...
  outputs:
   - backend: <name of backend object>
      input: <name of input we feed data to>
      expression: <expression for the value>
  	...
  variables:
   - name: <variable name>
      expression: <expression for the variable>
  	... 
  resources:
    ram: <expression for RAM>
    cpu: <expression for cores>
    replicas: <expression for replica count>
