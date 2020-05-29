When making a change to a project, there is a fairly typical sequence of
events. The aim of this document is to describe the steps, why they
exist, and possible alternatives to a developer who is not familiar with
the git based development sequence.

# Assumptions

The following steps assume that this is a project being managed using
git for source code control, publishing to a github repository (“repo”).
The specific tools used for editing, compilation, debug, test etc are
entirely a personal or project dictated choice. The project has already
been set up and all the necessary repositories have been cloned.

Also, the steps will be described based on a Windows environment but
there are equivalents for other environments and nothing is particularly
Windows specific.

# Preparation

Using a Windows Command line, navigate to the directory containing the
root of the local clone of the repository. Make sure the current branch
is set to “master” (or whatever branch the change is to be based on),
there are no outstanding changes and update the branch

  - git status

make sure there are no outstanding changes that are left over from
another change and assuming the current branch is NOT master, either
“commit” those changes or “stash” them as required.

  - git checkout master

switch to the master branch

  - git pull

update the contents of the local copy of the master branch to make sure
the new change will be based on the up to date contents of the master
branch

  - git checkout -b \<new-branch\>

create a new branch specifically for the change with an identifiable
name and switch to that branch

At this point, the branch is now ready to start making changes.

A common problem is to forget to switch to a new branch for the change
before making a set of edits. If so, a simple remedy is to perform the
“git checkout -b \<new-branch\>” without any commit. This effectively
transfers the edits to the new branch and removes them from the master
branch. If the “git pull” was also omitted, then after creating the new
branch, those edits can be committed to the branch and any updates to
the master branch can be merged into the change specific branch.

  - git add \<files for commit\>

  - git commit -m “prepare for merge”

  - git merge master

The merge may require that the developer resolve any merge conflicts
that might arise. If the merge becomes to complicated, it may be simpler
to abandon the current change, delete the branch and then re-start the
procedure, ensuring that the master branch is updated prior to making
any changes.

  - git restore \<file1\> \<file2\> etc

  - git checkout master

  - git branch -d \<new-branch\>

# Edit, Compile, Debug, Commit

Using any tools and utilities desired, make the requisite edits to the
relevant files, compile those changes and then debug the results.

For example, within the CloudChamber project which primarily uses Go, a
convenient method to implement the compile and debug steps is to write a
test module and use that to trigger a debug session for the code being
modified. That is, for a module names newcode.go, also provide a module
newcode\_test.go and in that module have a series of go functions named
TestXxxXxx() which make calls into the code in the newcode.go module.

Using an IDE such as VsCode, a breakpoint can be set on the specific
test function and then the IDE can be used to debug that test function
which will launch the debugger, which in turn will inherently compile
the necessary modules, and then stop at the breakpoint. In VsCode, each
TestXxxXxx() fun will have a pair of links just above the function
definition to either “run test” or “debug test”. Calling the code to be
tested from such a TestXxxXxx() function is an easy way to debug the
being written for the project.

Once the changes reach a convenient state, they can/should be checked-in
within the local copy of the repository, e.g.

  - git add \<file1\> \<file2\>

  - git add \<file3\>

  - git commit -m “\<comment describing included edits\>”

While developing the overall change, there may well be multiple passes
through the edit, compile, test, commit stages and committing the
changes can a useful method of tracking the completion of intermediate
states. There is no need to publish these changes to the main repository
at this time.

A quick check using

  - git status

is useful to ensure there are no outstanding files left over. If so,
they can be added and committed as necessary.

# Project Compile/Build

Many projects will have a mechanism to “build” the entire project. For
CloudChamber this is achieved by running the buildall script. To run
this script

  - pushd %GOPATH%\\src\\github.com\\Jim3Things\\CloudChamber\\build

  - buildall

The script will compile and link all the buildable files for the entire
project.

# Project Test

Many projects have a mechanism to run all the tests for the entire
project. For CloudChamber this is achieved by running the testall
script. To run this script

  - pushd %GOPATH%\\src\\github.com\\Jim3Things\\CloudChamber\\build

  - testall

The script will run all the test modules for all the components of the
project.

# Pre-Publish Verification

Verify that all the edits intended to be included within the overall
change have been committed to the local repository, i.e. “git status”
shows no outstanding uncommitted edits in the \<new-branch\> local
branch.

# Publish

Once the edits are ready, the branch \<new-branch\> should be
“published” to the main repository. This process will create a
branch in the main repository which will contain all the changes from
the matching branch in the local repository. This can be done from the
command line using a git command or frequently using a utility such as
“Github Desktop”.

Assuming GitHub Desktop is being used, and that the local repository has
been properly configured, on the bar under the menu options, it should
already be showing the current local repository and the current branch
(i.e. \<new-branch\>) in that repository. If there is a committed change
that has not yet been published, there will be an option to publish that
commit. By clicking on that button, all the commits in the current
branch in the local repository will be published to the main repository,
in a branch \<new-branch\> which matches the local branch name.

Note, assuming the changes were made and published from \<new-branch\>,
i.e. NOT the “master” branch, at this point, the change is in the main
repository, but are NOT yet part of the master branch.

# Pull Request

A pull request is a means of asking other team member to review and
comment on the change that is being proposed for pulling into the master
branch. When the Pull request (“PR”) is created, the author is given the
option to ask specific team members to review the change.

A PR can be created from the github.com site directly, or using the
“Github Desktop” application from the “Branch-\>Create Pull Request”
menu option which will automatically open a browser to the correct
location on the Github.com site with most of the fields automatically
filled in. Check the PR title is correct, add some explanatory text, and
select the reviewers to review the proposed change. When everything is
correct, click on the “Complete Pull Request” button.

# Code Review

Once the PR is created, Github will send a message to each of the
reviewers to notify them the proposed change is ready for review. The
reviews will add any comments they think appropriate, and then they can
either send the comments as is, request further changes or approve the
PR.

If further edits to the change are needed, the updates are made to the
\<new-branch\> branch in the local repository, built, tested and
verified as before and then committed to the local branch. Then these
further updates are added to the current PR by publishing them to the
main repository using the same process as for the initial publish, and
providing the branch names match i.e. from \<new-branch \> in the local
repository to \<new-branch\> to \<new-branch\> in the main repository,
the new updates will be added to the PR and the author can request a
re-review (using the github.com site).

# Merge

When sufficient team members have approved the change (The CloudChamber
project at present is aiming for 1 or more approvals), the author can
then use the “Squash and merge” button on the github.com site to
complete the PR and “pull” the changes into the master branch.

At this point, the change is complete.

Once the changes are merged into the master branch, an option will be
provided to delete the \<new-branch\> branch from the main repository.
If no further changes using that branch are anticipated, the typical
case, the branch can be deleted. Note this has no effect on the
\<new-branch\> in the local repository.

Back on the local repository, to verify the changes were properly
merged/pulled into the master branch

  - git status

to verify there are no outstanding lingering files, then

  - git checkout master

  - git pull

to switch to the current “master” branch and update that branch in the
local repository. The files which were in the change can be examined to
ensure all the expected updates are properly in place.

if the local topic branch \<new-branch\> is no longer required, it can
be deleted using

  - git checkout master

  - git branch -d \<new-branch\>

or possibly

  - git branch -D \<new-branch\>

was not completely merged with the current master branch, a likely
scenario.
