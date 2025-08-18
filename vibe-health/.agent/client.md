# Vibe Health Client

We are further refining the vibe health. Check the vibe-health/.agent/health-api-project.md to get up to speed.

Notice the stubbed .agent/client.md, we will be iterating on the design.

Write no code unless authorized. Our scope is strictly limited to requirements gathering, research, and writing the spec.

## Android Prototype

We will be starting with an Android prototype.

The user has an Android device to test on.
The user is developing the app on a Macbook.
The user is developing the app on a VSCode variant.

The user desires comprehensive research on prototyping tools available given this development setup.
* Can we have direct Android screen capture? Literally run android device connected to the local environment?
* Would this allow for the testing of push notifications in a local environment setup?
* Our goal is to have near zero workflow delays in testing changes on Android.

We will research, with sources, Android application development tools and ecosystem. Those tools will be added to this document for the user to review.

## Features

### Vibe Check

#### Push Notifications

Throughout the day, a set of push notifications will be pushed to user.

Morning: Sleep Check

```
How did you sleep?
😁     😐     😡     ➕
```

Each emoticon is a button that sends the value to the server.
The user is not directed to the application.
The + emoticon leads to the vibe entry feature in the application.

Afternoon: Mood Check

```
Hows the day going?

😊   🙁   😴💤   ➕
```

#### Vibe Checks in App

In the application the user can see todays vibes: sleep and mood.
They can select the vibe to enter a vibe entry screen.

Vibe entry screen consists of:
* Title (How did you sleep/how is day going?)
* Emoticon set
* Text body if user wants to add comments.
* Submit button (submit button only closes the entry, responses are saved automatically as theyre entered)

#### Vibe Backend & Data

* We need to processs the vibe features into a timeseries data object.
* The object is sent to the API and then in turn to the database(s).
* The vibes are to track sleep mood, day mood over time.
* This will pair with quantitative data like sleep score (garmin, etc.), nutrition and meals entered. etc.
  * This allows the user to see the relationship between sleep, nutrition, and excercise and their daily mood.
  * The hypothesis is better sleep, nutrition, and excercise will lead to better mood.

#### Vibe Data Model

The above feature raises interesting questions to consider about the nature of the data. My initial idea was to have a time series type database, something ingested into like InfluxDB as example.

However, I'm wondering if this DB type is flexible enough for the feature as written.
This is a relatively simple data object, emoticon (likely normalized to a DB friendly mood value).
But the editable nature of the object makes me wonder if timeseries data object the best approach.
Timeseries makes it easy to graph over time. But I assume the data type is inflexible.
I expect it would be easier to have a more flexible data type and graph that over time, then it is to have a timeseries DB object editable.