The following instructions have proven useful to setup a working
environment suitable for building and testing CloudChamber. An
assumption has been made that there is at least basic familiarity with
using Windows both from the desktop and via the Command Prompt.

The environment will make use of the following

  - Git

  - GitHub Desktop

  - Go

  - VsCode

  - protoc

VsCode is not required and other IDEs might be preferred. Configuring
those IDEs is left as an exercise for the reader.

# Assumptions

There are some paths required with associated environment variables.
Throughout the remainder of this documents these will be assumed to have
the following values

  - These instructions assume that the Go environment is rooted at the
    C:\\Chamber directory and that the value of the environment variable
    GOPATH is set to C:\\Chamber. This directory is an arbitrary choice
    and an alternative location on a local filesystem can be chosen
    instead of C:\\Chamber to suit the system being used provided GOPATH
    is set to the same location.

Note, avoid using a network share for the local copy of the repository
as some problems have been encountered for files which originated on a
non-Windows system.

  - These instructions install some tools and assume these are placed in
    the C:\\GoTools directory. This is an arbitrary selection and any
    alternative location can be selected to suite the system being used.
    If an alternative location is used, substitute the selected location
    for any reference to C:\\GoTools.

  - These instructions assume the Windows system is installed on the C:
    drive and standard installation locations are used. In particular,
    the instructions assume that the value of the standard environment
    variable ProgramFiles has the value “C:\\Program Files”. If an
    alternative value is used, substitute this value wherever the
    instructions refer to “C:\\Program Files”.

# Windows Basics

## Windows Command prompt

A Windows Command Prompt can be started using

> Start-\>Windows System-\>Command prompt

It is recommended that a desktop shortcut be set up or the Command
prompt be pinned to the taskbar. Pinning to the taskbar can be achieved
by right-clicking on any toolbar icon and then clicking “Pin to taskbar”
menu option.

## Examining Current Environment Variables

Within a Command prompt, type

> set \<varable\>

to see the variable (if defined) and its current value. The specified
variable name is treated as a prefix with an implicit wildcard and so
*all* variables with the prefix will be listed. If no prefix is
supplied, then all the currently defined variables will be listed.

## Defining or Modifying an Environment Variable

Within a Command prompt, a variable can be defined or changed using

> set \<variable\>=\<newValue\>

Variables are not case-sensitive. Note that everything following the ‘=’
is treated as part of the value, including any leading spaces.

## Un-Defining an Environment Variable

Within a Command prompt, a variable can be removed using

> set \<variable\>=

where there is nothing after the ‘=’. This will delete the variable
completely. It is generally a good idea to verify this with a

> set \<variable\>

to ensure the deletion took place as intended.

## Modifying Default System or User Environment Variables

Changes made using the “set” command directly within a Command prompt
are only effective within that Command prompt. If a permanent/persistent
change is required, the “setx” command can be used (see “setx /?” for
help). Alternatively, a variable can be added or modified using a dialog
box at

> Start-\>Settings-\>System-\>About-\>Related settings/System
> Info-\>Advanced system settings

This will bring up a dialog box. Click on the "Environment Variables"
button. This will bring up another dialog box with two panels. The upper
panel is the user specific variables, and the lower is the system wide
variables. Either can be changed (not both) but generally it is safer to
restrict changes to the user specific environment.

For example, to modify the default “PATH” environment variable, select
the "Path" variable and then click on the edit button. This will bring
up yet another dialog box. Click on the "New" button and in the
highlighted line edit the value. A change to add (say) the GOTOOLS
binary path, add the extra path to the end of the set. A value can use
another environment variable, provided ith was defined earlier in the
order than the variable currently being modified.

Click on OK for all the opened dialog boxes.

Note this will only affect the PATH environment variables in any Command
prompts or other applications started after change has been completed.
Anything started before will not see the new PATH.

Start a new Windows Command prompt and see if the new value has been
properly applied.

## Using the Current Value of an Environment Variable

To use the current value of an environment variable, enclose the
variable name in '%' characters. For example

> pushd %USERPROFILE%

When this line is encountered, the %USERPROFILE% is replaced by the
current value of the USERPROFILE environment variable and then in the
example above, the pushd command uses that expanded path.

This can be verified using any variable using a Windows Command prompt
by typing

> echo %USERPROFILE%
> 
> echo %USERNAME%

# Installing Git

## Verifying Git installation

If git is already installed, starting from a Windows command, type

> git

The utility should supply the top-level help. If not, git is probably
not correctly installed. Either a repair can be attempted, but it is
generally simpler to uninstall it and then re-install it (recommended).

To see if it can be simply repaired, at the Command prompt check the
value of the PATH environment variable verify that it contains (assuming
that the ProgramFiles environment variable has the value “C:\\Program
Files”)

> C:\\Program Files\\Git\\cmd

If not, the “C:\\Program Files\\Git\\cmd” directory needs to be added to
the PATH environment variable. As an experiment, add the path to the
current PATH variable using

> set PATH=”%PATH%;C:\\Program Files\\Git\\cmd”

and then try to run git again. If this works, then just updating the
default per-User environment settings may well suffice.

If not, or if preferred, just uninstall git and then re-install it.

## Installing Git

Git can be downloaded from

> https://git-scm.com/downloads

and it when the installed is executed

1.  Select Git from "Command Line…"

2.  Select "Use the native Windows Secure Channel library"

3.  Select "Checkout Windows-style, commit Unix-style line endings"

4.  Select "Use Windows default console window"

5.  Select
    
    1.  Enable File system caching
    
    2.  Enable Git credential manager
    
    3.  NOT enable symbolic links

# Installing GitHub Desktop

The GitHub desktop application can be downloaded from

> https://desktop.github.com/

## Adding a known repository

Once a repository has been cloned either via "git clone" or "go fetch"
you can add it using

> File-\>Add local Repository…

and then fill in the path to the local repo e.g. using the values from
below (assuming the GOPATH environment variable has the value
“C:\\Chamber”)

> C:\\Chamber\\src\\github.com\\Jim3Things\\CloudChamber

# Installing Go

The go package can be downloaded from

> https://golang.org/dl/

and then the package should be installed as normal. After installation,
using a newly started Command prompt, verify the installation by typing

> go

which if working, will display the top-level help for the go tool.

# Installing The protoc compiler

Download the pre-built binary package. This can be obtained from

> <https://developers.google.com/protocol-buffers/docs/downloads>

for the latest version (currently v3.11.4). Assuming the Windows
installation is 64bit (use Command prompt to examine the
PROCESSOR\_ARCHITECTURE environment variable) the required package will
be something like

> protoc-3.11.4-win64.zip

Create a directory where this should be installed, e.g. C:\\GoTools,
then copy the contents of the zip file to this directory. E.g.

> Bin\\protoc,exe ==\> C:\\GoTools\\bin\\protoc.exe
> 
> Include\\google ==\> C:\\GoTools\\include\\google

## Adding the Protoc Compiler to the PATH Environment Variable

Make sure the directory where the protoc compiler was placed is added to
the system PATH environment variable.

This can be done every time you start a Windows Command Prompt windows
using

> set PATH=%PATH%;C:\\GoTools\\bin

or you can change the persistent environment PATH variable (recommended)
using any of the previously described methods.

# Installing the Node.js Package Manager (NPM)

The CloudChamber UI portion of the project uses several features of a
set of frameworks which are installed using the NPM utility and then the
project itself is built using NPM.

## Installing the NPM Utility

Download the pre-built binary package. This can be obtained from

> <https://nodejs.org/en/>

for the latest version (currently 14.5.0). Run the installer and leave
the installation location to the default. The additional tools are not
required and the checkbox can be left un-selected.

# Setting up the Cygwin Utility Environment

The Cygwin environment provides a comprehensive suite of UNIX derived
utilities which are readily available on UNIX and UNIX like machines,
e.g. Linux, but which are not typically present on most Windows
machines.

At present, CloudChamber is only using the “make.exe” utility and then
only during package update and build. Later, more utilities may well be
needed.

## Installing the Base Toolset

The Cygwin installer “setup-x86\_64.exe” can be downloaded from
<https://www.cygwin.com>. Run the installer and respond to the prompts,
typically by accepting the defaults, e.g.

> Choose a Download Source:
> 
> Install from Internet
> 
> Select Root Install Directory:
> 
> Root Directory: c:\\cygwin64 (use supplied default unless override
> preferred)
> 
> Install For: All Users
> 
> Select Local Package Directory:
> 
> Local Package Directory: C:\\Users\\\<username\>\\Downloads (use
> supplied default unless override preferred)
> 
> Select Your Internet Connection:
> 
> Use System Proxy Settings
> 
> Choose a Download Site:
> 
> \<download site\>

This download site should be one that is geographically close to the
machine where the installation is taking place. For example, in the
Pacific NW, <http://mirrors.kernel.org> is convenient as that is located
in Portland, OR. For NZ, you might choose
<http://ucmirror.canterbury.ac.nz>

This will then present a new page "Cygwin Setup - Select Packages". With
the "View" drop-down menu set to full, in the search box, type "make".
This should cause a bunch of possible packages to be displayed.

Search down the list for the "make" package, and then in the "New"
column (probably the third column) select the highest available version
to be installed. Currently this is 4.3-1.

Click on the "Next" button and keep on selecting Ok, until the utility
starts the installation.

When the installation is complete, the utility will display the "Create
Icons" page. Ensure the options "Create Icon on Desktop" and "Add icon
to Start Menu" are both selected (unless otherwise preferred) and click
on the "Finish" button.

The utility should now create the icons and exit.

## Getting Utilities to Run from the Windows Command Line

By default, the Cygwin installed tools are only active in the Cygwin
environment, e.g. in the "Cygwin64 Terminal". To use the make utility
from within a standard Windows Command line, the default Cygwin path
needs to be added to the Windows PATH environment variable. To verify,
start a Windows Command line and then, assuming Cygwin was installed to
"C:\\cygwin64" type

> set PATH=c:\\cygwin64\\bin;%PATH%

to temporarily modify the environment and then

> make --version

where the version is preceded by a double ‘-‘ character, should cause
the make utility to display its version number.

If this is working correctly, either use this method to modify the PATH
environment variable each time a Windows Command line is going to be
used or modify the System or User environment variable if desired.

# Setting up the Initial CloudChamber Repository

## Fetch CloudChamber Itself

There are several ways to do this, such as using git, using go, or using
the GitHub desktop app. Select a preferred method.

Using a Command prompt, start by setting the value of the GOPATH
environment variable to the directory to be used as the root of the Go
project directory. This is assumed to be C:\\Chamber for these
instructions, but another directory could be used as an alternative.

> set GOPATH=C:\\Chamber

This will set the value temporarily but by using the methods described
above, this could be set permanently which would allow the setting to
persist across logout/login and reboots.

If using git, type either

> git clone <https://github.com/Jim3Things/CloudChamber>
> %GOPATH%\\src\\github.com\\Jim3Things\\CloudChamber

or if preferred, the command can be broken down into multiple steps
using

> mkdir %GOPATH%\\src\\github.com\\Jim3Things\\CloudChamber
> 
> pushd %GOPATH%\\src\\github.com\\Jim3Things
> 
> git clone <https://github.com/Jim3Things/CloudChamber>

This will initialize a git repository and populate it with the
CloudChamber files under the specified directory.

If using the go tool to fetch the initial copy of the repository, type

> go get github.com/Jim3Things/CloudChamber

If using the GitHub desktop application, remember to add this newly
created repository to the set of known repositories.

## Fetch Packages Used by CloudChamber

Once the CloudChamber repo is available, fetch the remaining packages

> cd /d
> %GOPATH%\\src\\github.com\\Jim3Things\\CloudChamber\\build\\dev\_tools
> 
> fetchall

This will fetch all the packages needed by the CloudChamber packages,
install the protoc Go support and protobuf validation modules.

A number of errors of the form

> Can’t load package: package \<PackageName\>: no Go files in
> \<PackagePath\>

may occur, particularly for the packages

  - github.com/golang/protobuf

  - go.opentelemetry.io/otel

  - github.com/Jim3Things/CloudChamber

If so, these can be safely ignored.

## Build CloudChamber

Once all the packages and tools are available (i.e. after fetchall has
run), the CloudChamber package can be built by typing

> cd /d %GOPATH%\\src\\github.com\\Jim3Things\\CloudChamber\\build
> 
> buildall

This should build all the CloudChamber services and copy the generated
executables and support files to

> %GOPATH%\\src\\github.com\\Jim3Things\\CloudChamber\\deployments

# Setting up the Initial CloudChamber UI Repository

## Fetch the CloudChamber UI Itself

This is just another GitHub project so any of the normal methods can be
used to set up the git repository, such as

> pushd %GOPATH%\\src\\github.com\\Jim3Things
> 
> go get github.com\\Jim3Things\\cloud\_chamber\_react\_ts

If using the GitHub desktop application, remember to add this newly
created repository to the set of known repositories.

## Install the additional packages using NPM

To prepare the additional packages to be used with the UI project, at a
Windows command line

> cd /d %GOPATH%\\src\\github.com\\Jim3Things\\cloud\_chamber\_react\_ts
> 
> npm install react --save
> 
> npm install react-dom --save
> 
> npm install @material-ui/core
> 
> npm install @material-ui/icons
> 
> npm install @material-ui/lab

The “--save” option use a double ‘-‘ character)

Note that after an individual step, there may be a message suggesting
that there are some potential security issues. If so, run the suggested

> npm audi fix

command.

## Build The CloudChamber UI Project

Once the repository has been populated, and the additional node.js
packages installed, the project can be built using the following

> cd /d %GOPATH%\\src\\github.com\\Jim3Things\\cloud\_chamber\_react\_ts
> 
> npm run-script build

# Installing the VsCode Application

## Useful References for VsCode

The following reference may prove useful

  - <https://github.com/Microsoft/vscode-go/wiki/Go-tools-that-the-Go-extension-depends-on>

  - https://github.com/microsoft/vscode-go\#go-language-server

## VsCode Installation Package

The VsCode installer can be found at

> https://code.visualstudio.com/download

and then installed by executing the installer.

## Installing the VsCode Go support

Start VsCode. Using extension list, (5th icon down on leftmost edge - a
box marked in quarters with one displaced) or menu View-\>Extensions,
search for the Go extension

> Go - Rich Go language support for Visual Studio Code

from Microsoft. Install this extension.

## Configure the Go extension

Once the extension is installed, a few values need to be configured.
Start VsCode and navigate to the settings panel at

> File -\> Preferences -\> Settings

In displayed panel navigate to the Go specific configurable items at

> Extension-\>Go

Scroll the right hand panel down to locate the GoPath settings and click
on "Edit in settings.json”. Edit the value for go.gopath to match the
value for the GOPATH environment variable. Note that the directory
separator character ‘\\’ needs to be escaped with another ‘\\’
character, so for a GOPATH environment variable value of “C:\\Chamber”,
the go.gopath value should be “C:\\\\Chamber”. Save the value and close
the settings.json tab.

Once the GoPath value has been set, scroll the settings panel to “Infer
GoPath” and enable.

Scroll panel to “Tools GoPath” and set to directory where other binaries
were installed, e.g. C:\\GoTools.

Scroll panel to Use Language Server and enable

To confirm the settings, in any of the options, click on "Edit in
settings.json" and confirm the layout is something like

> {
> 
> "go.toolsGopath": "C:\\\\GoTools",
> 
> "go.inferGopath": true,
> 
> "go.gopath": "C:\\\\Chamber",
> 
> "go.useLanguageServer": true,
> 
> }

There might also be other settings, but they can be ignored for now.

Close the settings.json panel/tab

In the main VsCode window, note bottom right, where it says analysis
tool missing. Click on and select install.

Restart VsCode

If VsCode complains about lack of language server (pop-up in the bottom
right corner), click on install.

## Installing the VsCode Go debugging support

In one of the CloudChamber test source code files, e.g. users\_test.go,
locate the TestMain() function and click on the “debug test” link just
above the declaration of the TestMain() function.

If the debugger support is not currently installed, there will be a
pop-up to say dlv is missing. Click on install.

Once installed, click on the “debug test” link just above the TestMain()
declaration and check the output of the test in the “Debug Console”
window below the source code window. The output should conclude with a
PASS message with a process exit code of 0.

## Using the Debugging Support

The debugger follows much of the normal Microsoft debugger command keys
(compare with Visual Studio debugger, WinDbg etc), e.g.

  - F9 - toggle breakpoint

  - F10 - single step

  - F11 - step into (a function)

  - F5 - run

At this point, everything needed to use the VsCode editor/tool-set
should be working. However, to use the Windows Command prompt to fetch
and build all the needed packages we will also need to install the
protobuf compiler (protoc) and the associated validator.

## Installing the VsCode Protobuf Support (Optional)

Start VsCode

Using extension list, search for the vscode-proto3 extension

Install the extension

# Checking Everything is working 

Start a Windows Command Prompt / Git Command prompt and get to the
CloudChamber repository. Type

> set GOPATH

and verify the displayed variable is properly set to the root of the
repository. Type

> set PATH

and verify the displayed path contains the "GoTools\\bin" directory.

Build CloudChamber and check that the

> %GOPATH%\\github.com\\Jim3Things\\CloudChamber\\deployments

contains the generated executables for controllerd, inventoryd,
sim\_supportd and web\_server, a README.md, a cloudchamber.yaml
configuration file and some cmd files to start various components.
