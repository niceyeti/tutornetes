# AWS CDK notes for the learning of all the things


### Best Practices

With the AWS CDK, developers or administrators can define their cloud infrastructure by using a supported programming language. CDK applications should be organized into logical units, such as API, database, and monitoring resources, and optionally have a pipeline for automated deployments. The logical units should be implemented as constructs including the following:

* Infrastructure (such as Amazon S3 buckets, Amazon RDS databases, or an Amazon VPC network)
* Runtime code (such as AWS Lambda functions)
* Configuration code

Interesting:
```
"Although it's possible, we don't recommend putting multiple applications in the same repository, especially when using automated deployment pipelines. Doing this increases the "blast radius" of changes during deployment. When there are multiple applications in a repository, changes to one application trigger deployment of the others (even if the others haven't changed). Furthermore, a break in one application prevents the other applications from being deployed."
```

Infrastructure and runtime code live in the same package:
```
"In addition to generating AWS CloudFormation templates for deploying infrastructure, the AWS CDK also bundles runtime assets like Lambda functions and Docker images and deploys them alongside your infrastructure. This makes it possible to combine the code that defines your infrastructure and the code that implements your runtime logic into a single construct. It's a best practice to do this. These two kinds of code don't need to live in separate repositories or even in separate packages."
```

Use deployment-time cyclomatic complexity sparingly: if, conditions, and parameters in CloudFormation.
Instead, try to make all decisions, such as which construct to instantiate, in your AWS CDK application by using your programming language's if statements and other features. 

#### Testing

* Fine-grained assertions test specific aspects of the generated AWS CloudFormation template, such as "this resource has this property with this value." These tests can detect regressions. They're also useful when you're developing new features using test-driven development. (You can write a test first, then make it pass by writing a correct implementation.) Fine-grained assertions are the most frequently used tests.

* Snapshot tests test the synthesized AWS CloudFormation template against a previously stored baseline template. Snapshot tests let you refactor freely, since you can be sure that the refactored code works exactly the same way as the original. If the changes were intentional, you can accept a new baseline for future tests. However, CDK upgrades can also cause synthesized templates to change, so you can't rely only on snapshots to make sure that your implementation is correct.



### Glossary

#### Resources

Consider starting with testing, direct grok path: https://docs.aws.amazon.com/cdk/v2/guide/testing.html

A good off and running reference, no filler: https://medium.com/contino-engineering/increase-your-aws-cdk-lambda-development-speed-by-testing-locally-with-aws-sam-48a70987515c


#### IAM

* Principals: An IAM principal is an authenticated AWS entity representing a user, service, or application that can call AWS APIs. The AWS Construct Library supports specifying principals in several flexible ways to grant them access your AWS resources.
* Grants: grants are like bindings. Every construct that represents a resource that can be accessed, such as an Amazon S3 bucket or Amazon DynamoDB table, has methods that grant access to another entity. All such methods have names starting with grant (grantRead, grantWrite).
* Roles: represents IAM roles, containing policies denying/allowing actions upon things.
```
role.add_to_policy(iam.PolicyStatement(
    effect=iam.Effect.DENY,
    resources=[bucket.bucket_arn, other_role.role_arn],
    actions=["ec2:SomeAction", "s3:AnotherAction"],
    conditions={"StringEquals": {
        "ec2:AuthorizedService": "codebuild.amazonaws.com"}}
))
```

#### Context

Runtime context is defined in key-vals passed to your app via json or the cmd line.

The core concept behind context is caching: the CDK Toolkit uses context to cache values retrieved from your AWS account during synthesis. Because these values are provided by your AWS account, they can change
between runs of your CDK application. This makes them a potential source of unintended change. The
CDK Toolkit's caching behavior "freezes" these values for your CDK app until you decide to accept the
new values.

Contexts provide stability against available features, changes, and versions of things like AMI images.

#### Aspects

Aspects are a way to apply an operation to all constructs in a given scope, modifying them or adding tags.

Somehow these can be used to enforce service control policies and permission boundaries, ie assertions about organizational security policies.


#### Bootstrapping
Bootstrapping is the process of provisioning resources for the AWS CDK before you can deploy AWS CDK apps into an AWS environment. (An AWS environment is a combination of an AWS account and Region).

#### Abstraction levels
The AWS CDK lets you describe AWS resources using constructs that operate at varying levels of abstraction.

* Layer 1 (L1) constructs directly represent AWS CloudFormation resources as defined by the CloudFormation specification. These constructs can be identified via a name beginning with "Cfn," so they are also referred to as "Cfn constructs." If a resource exists in AWS CloudFormation, it exists in the CDK as a L1 construct.

* Layer 2 (L2) or "curated" constructs are thoughtfully developed to provide a more ergonomic developer experience compared to the L1 construct they're built upon. In a typical CDK app, L2 constructs are usually the most widely used type. Often, L2 constructs define additional supporting resources, such as IAM policies, Amazon SNS topics, or AWS KMS keys. L2 constructs provide sensible defaults, best practice security policies, and a more ergonomic developer experience.

* Layer 3 (L3) constructs or patterns define entire collections of AWS resources for specific use cases. L3 constructs help to stand up a build pipeline, an Amazon ECS application, or one of many other types of common deployment scenarios. Because they can constitute complete system designs, or substantial parts of a larger system, L3 constructs are often "opinionated." They are built around a particular approach toward solving the problem at hand, and things work out better when you follow their lead.

I don't quite understand, but objects of each layer can be created from one another, for a desired level of abstraction, or to encode certain organizational interests.

#### AWS Infrastructure as Code

It is code. As Infrastructure. That is code. As such. QED.

#### Fargate
Like EC2 except it allows you to use containers as the fundamental compute primitive.
They are also priced on-demand, and thus you pay only for the transactional costs of use, not idle operation time.
* sec reqs
* IAM and roles
* networking interfaces
* compute resources

Example:
```
aws ecs run-task --launch-type FARGATE --cluster BlogCluster --task-definition blog --network-configuration "awsvpcConfiguration={subnets=[subnet-b563fcd3]}"
```

Responsibilities:
-> task -> task group -> cluster -> 

#### Pipenv

Combines pip and venv into a single tool.

Usage:
1. Install and create a new venv: `pipenv install`
2. Uninstall: `pipenv uninstall` some dep
3. Generate pip-lock file: `pipenv lock`
4. `check`: look up security vulnerabilities
5. `graph`: show the dependency graph
6. `shell`: spawn a shell with venv activated
7. `run`: run a command with the venv activated, e.g. `pipenv run pip3 freeze`.

#### Stacks

A Stack is the unit of deployment on ECS, and presumably the unit of a billable workload (or close to it).

```
app = App()
MyFirstStack(app, "first stack")
MySecondStack(app, "second stack")
app.synth()
```
Now view these stacks in the app:
* `cdk list`
    ```
    first stack
    second stack
    ```
#### Environment

Defined by the account and region to which a stack will deploy.

#### Resources

Constructs, aka various AWS resources such as SQS queues, and so forth.

Pattern: `s3.Bucket(self, "MyBucket")`

#### Identifiers

These must be unique within the scope they are used.

#### Paths

Constructs in the App class form a hierarchy; their ids form a path.
Unique ids are assigned by suffixing 8-digit hashes to each path component, thus distinguishing A/B/C from A/BC.

#### Logical ID

A path component (a Resource) suffixed with an 8-digit value form the logical id of a resource: ABC-12345678.


#### Tokens

Tokens are an implicit part of string identifiers; be wary of modifying strings, and aware of the data types. I think I will only
learn this through coding.

#### Parameters

AWS CloudFormation templates can contain parameters whose value is only supplied/known at deployment time. Provided via the command line in the cdk environment:
```
cdk deploy MyStack --parameters uploadBucketName=uploadbucket
```
However, note that params are something of an anti-pattern:
```
Generally, it's better to have your CDK app accept necessary information in a well-defined way and use it directly to declare constructs in your CDK app. An ideal AWS CDK-generated AWS CloudFormation template is concrete, with no values remaining to be specified at deployment time. 
```
#### Tagging

Tags are informational key-value elements that you can add to constructs in your AWS CDK app. A tag applied to a given construct also applies to all of its taggable children. They seem similar to k8s annotations, in that they can be used to categorize resources logically:
* Simplifying management
* Cost allocation
* Access control
* Any other purposes that you devise (see [Tagging Best Practices](https://d1.awsstatic.com/whitepapers/aws-tagging-best-practices.pdf))

#### Assets

Assets are local files, directories, or Docker images that can be bundled into AWS CDK libraries and apps. For example, an asset might be a directory that contains the handler code for an AWS Lambda function. Assets can represent any artifact that the app needs to operate.

You add assets through APIs that are exposed by specific AWS constructs. When you refer to an asset in your app, the cloud assembly that's synthesized from your application includes metadata information with instructions for the AWS CDK CLI. The instructions include where to find the asset on the local disk and what type of bundling to perform based on the asset type, such as a directory to compress (zip) or a Docker image to build.

* Amazon S3 assets: These are local files and directories that the AWS CDK uploads to Amazon S3.
* Docker Image: These are Docker images that the AWS CDK uploads to Amazon ECR.

Assets can incorporate RBAC through IAM:

```
asset = Asset(self, "MyFile",
    path=os.path.join(dirname, "my-image.png"))

    group = iam.Group(self, "MyUserGroup")
    asset.grant_read(group)
```










