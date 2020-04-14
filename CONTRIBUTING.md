# Introduction

## Contributing to Harsh

You're here to help? Awesome! Please feel welcoem and read the following sections to get a feel for the lay of the land, know how to ask questions, and how to work on something.

All contributors are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md). Please make sure you are welcome and friendly. Be kind. Be respectful. (BKBR). Be nice. It's a tough world out there.

I'm pretty much happy for any contribution though since the idea is to keep the applicaiton minimalist new feature requests will be deliberated before addiiton. Code improvements, bug fixes, tests, documentation, translations, and any sort of unambiguous improvement is most welcome. Please [submit a PR](#pull-requests)

## Get in touch

    Report bugs, suggest features in the Issues on Github and view source on GitHub.
    I'm most easily pinged on twitter via [@awws](https://twitter.com/awws)
    For email, please look at the [About page on Daryl's blog](https://daryl.wakatara.com/about)

## Using the issue tracker

The issue tracker is the preferred channel for [bug reports](#bugs),
[features requests](#features) and [submitting pull
requests](#pull-requests).

### Bug reports

A bug is a _demonstrable problem_ that is caused by the code in the repository.
Good bug reports are extremely helpful.

Guidelines for bug reports:

1. **Use the GitHub issue search** &mdash; check if the issue has already been
   reported.

2. **Check if the issue has been fixed** &mdash; try to reproduce it using the
   latest `master` branch in the repository.

3. **Isolate the problem** &mdash; ideally create a reduced test case.

A good bug report shouldn't leave others needing to chase you up for more
information. Please try to be as detailed as possible in your report. What is
your environment? What steps will reproduce the issue? What OS experiences the
problem? What would you expect to be the outcome? All these details will help
people to fix any potential bugs.

Example:

> Short and descriptive example bug report title
>
> A summary of the issue and the browser/OS environment in which it occurs. If
> suitable, include the steps required to reproduce the bug.
>
> 1. This is the first step
> 2. This is the second step
> 3. Further steps, etc.
>
> `<url>` - a link to the reduced test case
>
> Any other information you want to share that is relevant to the issue being
> reported. This might include the lines of code that you have identified as
> causing the bug, and potential solutions (and your opinions on their
> merits).

### Feature requests

Feature requests are welcome. But take a moment to find out whether your idea
fits with the scope and aims of Harsh. It's up to *you* to make a strong
case to convince the project's developer(s) of the merits of this feature. Please
provide as much detail and context as possible.

### Pull requests

Good pull requests - patches, improvements, new features - are a fantastic
help. They should remain focused in scope and avoid containing unrelated
commits.

**Please ask first** before embarking on any significant pull request (e.g.
implementing features, refactoring code), otherwise you risk spending a lot of
time working on something that the project's developers might not want to merge
into the project.


# Ground Rules
### Set expectations for behavior (yours, and theirs).
This includes not just how to communicate with others (being respectful, considerate, etc) but also technical responsibilities (importance of testing, project dependencies, etc). Mention and link to your code of conduct, if you have one.

> Responsibilities
> * Ensure cross-platform compatibility for every change that's accepted. Windows, Mac, and Ubuntu Linux.
> * Ensure that code that goes into core meets all requirements in this checklist: https://gist.github.com/audreyr/4feef90445b9680475f2
> * Create issues for any major changes and enhancements that you wish to make. Discuss things transparently and get community feedback.
> * Don't add any classes to the codebase unless absolutely needed. Err on the side of using functions.
> * Keep feature versions as small as possible, preferably one new feature per version.
> * Be welcoming to newcomers and encourage diverse new contributors from all backgrounds.



### Bonus points: Add a link to a resource for people who have never contributed to open source before.
Here are a couple of friendly tutorials you can include: http://makeapullrequest.com/ and http://www.firsttimersonly.com/

> Working on your first Pull Request? You can learn how from this *free* series, [How to Contribute to an Open Source Project on GitHub](https://egghead.io/series/how-to-contribute-to-an-open-source-project-on-github).

As a side note, it helps to use newcomer-friendly language throughout the rest of your document. Here are a couple of examples from [Active Admin](https://github.com/activeadmin/activeadmin/blob/master/CONTRIBUTING.md):

>At this point, you're ready to make your changes! Feel free to ask for help; everyone is a beginner at first :smile_cat:
>
>If a maintainer asks you to "rebase" your PR, they're saying that a lot of code has changed, and that you need to update your branch so it's easier to merge.

# Getting started
### Give them a quick walkthrough of how to submit a contribution.
How you write this is up to you, but some things you may want to include:

* Let them know if they need to sign a CLA, agree to a DCO, or get any other legal stuff out of the way
* If tests are required for contributions, let them know, and explain how to run the tests
* If you use anything other than GitHub to manage issues (ex. JIRA or Trac), let them know which tools they’ll need to contribute

>For something that is bigger than a one or two line fix:

>1. Create your own fork of the code
>2. Do the changes in your fork
>3. If you like the change and think the project could use it:
    * Be sure you have followed the code style for the project.
    * Sign the Contributor License Agreement, CLA, with the jQuery Foundation.
    * Note the jQuery Foundation Code of Conduct.
    * Send a pull request indicating that you have a CLA on file.

[source: [Requirejs](http://requirejs.org/docs/contributing.html)] **Need more inspiration?** [1] [Active Admin](https://github.com/activeadmin/activeadmin/blob/master/CONTRIBUTING.md#1-where-do-i-go-from-here) [2] [Node.js](https://github.com/nodejs/node/blob/master/CONTRIBUTING.md#code-contributions) [3] [Ember.js](https://github.com/emberjs/ember.js/blob/master/CONTRIBUTING.md#pull-requests)

### If you have a different process for small or "obvious" fixes, let them know.

> Small contributions such as fixing spelling errors, where the content is small enough to not be considered intellectual property, can be submitted by a contributor as a patch, without a CLA.
>
>As a rule of thumb, changes are obvious fixes if they do not introduce any new functionality or creative thinking. As long as the change does not affect functionality, some likely examples include the following:
>* Spelling / grammar fixes
>* Typo correction, white space and formatting changes
>* Comment clean up
>* Bug fixes that change default return values or error codes stored in constants
>* Adding logging messages or debugging output
>* Changes to ‘metadata’ files like Gemfile, .gitignore, build scripts, etc.
>* Moving source files from one directory or package to another

[source: [Chef](https://github.com/chef/chef/blob/master/CONTRIBUTING.md#chef-obvious-fix-policy)] **Need more inspiration?** [1] [Puppet](https://github.com/puppetlabs/puppet/blob/master/CONTRIBUTING.md#making-trivial-changes)

# How to report a bug
### Explain security disclosures first!
At bare minimum, include this sentence:
> If you find a security vulnerability, do NOT open an issue. Email XXXX instead.

If you don’t want to use your personal contact information, set up a “security@” email address. Larger projects might have more formal processes for disclosing security, including encrypted communication. (Disclosure: I am not a security expert.)

> Any security issues should be submitted directly to security@travis-ci.org
> In order to determine whether you are dealing with a security issue, ask yourself these two questions:
> * Can I access something that's not mine, or something I shouldn't have access to?
> * Can I disable something for other people?
>
> If the answer to either of those two questions are "yes", then you're probably dealing with a security issue. Note that even if you answer "no" to both questions, you may still be dealing with a security issue, so if you're unsure, just email us at security@travis-ci.org.

[source: [Travis CI](https://github.com/travis-ci/travis-ci/blob/master/CONTRIBUTING.md)] **Need more inspiration?** [1] [Celery](https://github.com/celery/celery/blob/master/CONTRIBUTING.rst#security) [2] [Express.js](https://github.com/expressjs/express/blob/master/Security.md)

### Tell your contributors how to file a bug report.
You can even include a template so people can just copy-paste (again, less work for you).

> When filing an issue, make sure to answer these five questions:
>
> 1. What version of Go are you using (go version)?
> 2. What operating system and processor architecture are you using?
> 3. What did you do?
> 4. What did you expect to see?
> 5. What did you see instead?
> General questions should go to the golang-nuts mailing list instead of the issue tracker. The gophers there will answer or ask you to file an issue if you've tripped over a bug.

[source: [Go](https://github.com/golang/go/blob/master/CONTRIBUTING.md#filing-issues)] **Need more inspiration?** [1] [Celery](https://github.com/celery/celery/blob/master/CONTRIBUTING.rst#other-bugs ) [2] [Atom](https://github.com/atom/atom/blob/master/CONTRIBUTING.md#reporting-bugs) (includes template)

# How to suggest a feature or enhancement
### If you have a particular roadmap, goals, or philosophy for development, share it here.
This information will give contributors context before they make suggestions that may not align with the project’s needs.

> The Express philosophy is to provide small, robust tooling for HTTP servers, making it a great solution for single page applications, web sites, hybrids, or public HTTP APIs.
>
> Express does not force you to use any specific ORM or template engine. With support for over 14 template engines via Consolidate.js, you can quickly craft your perfect framework.

[source: [Express](https://github.com/expressjs/express#philosophy)] **Need more inspiration?** [Active Admin](https://github.com/activeadmin/activeadmin#goals)

### Explain your desired process for suggesting a feature.
If there is back-and-forth or signoff required, say so. Ask them to scope the feature, thinking through why it’s needed and how it might work.

> If you find yourself wishing for a feature that doesn't exist in Elasticsearch, you are probably not alone. There are bound to be others out there with similar needs. Many of the features that Elasticsearch has today have been added because our users saw the need. Open an issue on our issues list on GitHub which describes the feature you would like to see, why you need it, and how it should work.

[source: [Elasticsearch](https://github.com/elastic/elasticsearch/blob/master/CONTRIBUTING.md#feature-requests)] **Need more inspiration?** [1] [Hoodie](https://github.com/hoodiehq/hoodie/blob/master/CONTRIBUTING.md#feature-requests) [2] [Ember.js](https://github.com/emberjs/ember.js/blob/master/CONTRIBUTING.md#requesting-a-feature)

# Code review process
### Explain how a contribution gets accepted after it’s been submitted.
Who reviews it? Who needs to sign off before it’s accepted? When should a contributor expect to hear from you? How can contributors get commit access, if at all?

> The core team looks at Pull Requests on a regular basis in a weekly triage meeting that we hold in a public Google Hangout. The hangout is announced in the weekly status updates that are sent to the puppet-dev list. Notes are posted to the Puppet Community community-triage repo and include a link to a YouTube recording of the hangout.
> After feedback has been given we expect responses within two weeks. After two weeks we may close the pull request if it isn't showing any activity.

[source: [Puppet](https://github.com/puppetlabs/puppet/blob/master/CONTRIBUTING.md#submitting-changes)] **Need more inspiration?** [1] [Meteor](https://meteor.hackpad.com/Responding-to-GitHub-Issues-SKE2u3tkSiH ) [2] [Express.js](https://github.com/expressjs/express/blob/master/Contributing.md#becoming-a-committer)

# Community
If there are other channels you use besides GitHub to discuss contributions, mention them here. You can also list the author, maintainers, and/or contributors here, or set expectations for response time.

> You can chat with the core team on https://gitter.im/cucumber/cucumber. We try to have office hours on Fridays.

[source: [cucumber-ruby](https://github.com/cucumber/cucumber-ruby/blob/master/CONTRIBUTING.md#talking-with-other-devs)] **Need more inspiration?**
 [1] [Chef](https://github.com/chef/chef/blob/master/CONTRIBUTING.md#-developer-office-hours) [2] [Cookiecutter](https://github.com/audreyr/cookiecutter#community)

# BONUS: Code, commit message and labeling conventions
These sections are not necessary, but can help streamline the contributions you receive.

### Explain your preferred style for code, if you have any.

**Need inspiration?** [1] [Requirejs](http://requirejs.org/docs/contributing.html#codestyle) [2] [Elasticsearch](https://github.com/elastic/elasticsearch/blob/master/CONTRIBUTING.md#contributing-to-the-elasticsearch-codebase)

### Explain if you use any commit message conventions.

**Need inspiration?** [1] [Angular](https://github.com/angular/material/blob/master/.github/CONTRIBUTING.md#submit) [2] [Node.js](https://github.com/nodejs/node/blob/master/CONTRIBUTING.md#step-3-commit)

### Explain if you use any labeling conventions for issues.

**Need inspiration?** [1] [StandardIssueLabels](https://github.com/wagenet/StandardIssueLabels#standardissuelabels) [2] [Atom](https://github.com/atom/atom/blob/master/CONTRIBUTING.md#issue-and-pull-request-labels)
