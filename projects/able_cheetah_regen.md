# ABLE Cheetah Regen

ABLE and Cheetah are a perrenial goal to get up and running and displaying again.
The whole stack could be automated with data pipelines, similar to slackyderm.

Question: is this the signals app we never built? For example, this application could be
conceived of as a single code file hosted in a web editor: specify a data source, and an indication (red/green)
is given that it is available. Likewise other functions could be integrated as well.
Basically the programming language would look like python but would decompose to underlying
constructs: a data store would be a live item (red/green status, schema info), analysis
code could ultimately resolve to a serverless function a la kubeless. The code file
would essentially define the entire application logic or use-case; the underlying components
would resolve what/how things run. This is the 'computer' to someone's 'solution'; I always
favored the computational view, never liked the feature-bloated 'solution'.

The benefit is that this could provide:
* review of python and ML libs
* data pipeline experience, kubeless, additional kubernetes workflows
* restoration of Cheetah/ABLE

Gist:
* Crawlers as CronJobs that periodically check for new content on sites, archives
* Pipelines that transform and persist this data
* CronJob Analyzers that consume the persisted data
* Analysis views
* Search apis
In sum, an end to end MLOps system for linguistic data. The goal would be to generalize as much as possible
for various intelligence/analysis apps for discrete data (natural language, social media, etc.)

The most generic problem statement and its components: a kubernetes-based MLOps NLP analysis platforms.
* Data sources: things you want to monitor (sites, apis, sql databases, confluence)
* Crawlers and transformers: Jobs that run and extract/transform this data
* Persistence/volumes: raw persisted data, or other output
* Analyzers: these monitor data volumes/versions for new data. On arrival, they re-run on some schedule/trigger.
  They export their results to some location. Analyzers are candidates for kubeless/serverless.
* Views: these are http views which consume versioned results from the analyzers.
* Recall the analysis engine underway and since cancelled at ***. The goal here is a similar stateful,
  observable-driven analysis pipeline: touch some upstream data, and the entire workflow triggers and re-runs.

The platform addresses the needs of analyses of discrete, periodic, versioned data and analytics.
The inputs and outputs should integrate with other systems, such that one's analyses could be hosted on other secure sites.

User stories:
1) I am Bob, and I want to monitor a major region/event. I specify some data sources, have some extractors/crawlers written, and analyses, and expect my system to actively monitor events and data.
    * sources, sentiment, potential group activity, other signals
2) I am Lisa, a data scientist. I want to access the data (raw or clean/transformed) for custom, journal-based analyses.
    * Recall the signals-analysis app idea. This is that, but for content/language data.
3) I am Commander Sig, a ground-operational leader. I need to extract and store a significant amount of speech
   data for further translation and analysis across the duration of a deployment in a particular theatre.
   This will provide useful analyses and queries of the development and resolution of a conflict.








