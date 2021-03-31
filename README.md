# Name of Project

The first phrase should describe the project briefly, for example, this is the main template to be used by all RD Github Repositories. 

Second paragraph is dedicated to describe it: What this library/api/whatever does and what does not. 
What the advantagens of this project? Why use it instead another public library. For example, 
by using this api, we abstract what log library is used in order to provide a simple interface 
to developers that wants to log into pre-defined formats.  

## Audience

Describe here the main audience from this project. Eg: This Template is designed for Developers or Tech Leaders. 

## Useful terminology

Lists definitions of terms that the reader needs to know to follow the tutorial. For example: what is lead? what is a conversion? You include charts here

![pretty-diagram](https://user-images.githubusercontent.com/18356186/54356236-e163b900-4639-11e9-9bb1-e171bcd2a025.png)

## Getting started

### Requirements

Lists concepts the reader should be familiar with prior to starting, as well as any software or hardware requirements. 
If possible, please create a link to instalation documentation or use it inline.

* [rd-docker installed](https://oraculo.rdstation.com.br/referencias/wiki/como-configurar-o-ambiente-de-desenvolvimento-utilizando-docker)
* Ruby v5.1.7 installed
* Bundle v2.0.1 installed
* Any of my dependencies up and running:

```bash
$ gem install my_dep_here
```

### Running in Local Environment

Provide description how to run locally but also command lines:

1. Start Container
```bash
$ rd-docker s
```
2. Access http://localhost:8008

3. Login using default credentials:
* Username: my_user
* Password: my_pass

### Running Tests

Provide description how to run tests locally. Command lines are really important:

```bash
$ rd-docker c
$ rspec .
```

### Runing in Production environment

Explains how to run in production environment and apply your changes too.

1. Access [Spinnaker](https://spinnaker.rdops.systems/#/applications/my-app/clusters);
2. Go to Pipelines on left menu. Click on em `Start Manual Execution` on disered execution

<img src="https://user-images.githubusercontent.com/9935397/82076477-48548600-96b4-11ea-8a13-84e14f6463b0.png" height="300">

3. Choose the branch

<img src="https://user-images.githubusercontent.com/9935397/82076681-979ab680-96b4-11ea-948f-974a3d518378.png" height="300">

4. Click on Run. 
5. Wait until finished
6. Test it using [Production URL](https://www.google.com)
7. Click on Continue to merge it

## What's next (Optional)

* Bullet points
* That you believe
* Are the next steps
* But don't try to predict all your future
