Logging made available through a global variable called logger.

* all logs should begin with a lowercase word etc
* log intent as soon as possible
* but keep normal output to a mimimum
* absolutely no private data should be logged ever

* all logs should decribe the behaviour of the current function, not describe what is about to happen
* EXCEPT for when making calls to core functions which do not do logging

* info level used during normal execution to give essential feedback to user
* info only used from within cli functions

* warns should be used to suggest user error, e.g operating on input files that do not exist
* the controller should log a warn and then return nil as the error

* debug used to helping user and developers troubleshoot issues
* debug used in cli functions and controllers
* debug should include context and ids

* trace output is assumed to be formatted with function name, file name and number
* trace used at the start of fuction and should include parameters
* trace can be used at the end of a function and include an output
