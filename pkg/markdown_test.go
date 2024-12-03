package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMarkdown(t *testing.T) {
	t.Run("panic on broken", func(t *testing.T) {
		rawContent := "+++\n???"

		// execute
		content, err := ParseMarkdown(rawContent)
		require.Error(t, err)

		// verify
		assert.Empty(t, content)
	})

	type args struct {
		rawContent string
	}
	tests := []struct {
		name string
		args args
		want Content
	}{
		{
			name: "title-only",
			args: args{
				rawContent: `+++
title = "Prepare"
+++`,
			},
			want: Content{
				Title: "Prepare",
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					SectionTitles: []string{},
				},
			},
		},
		{
			name: "state-only",
			args: args{
				rawContent: `+++
state = "incomplete"
+++`,
			},
			want: Content{
				Title:  "",
				State:  Incomplete,
				Weight: "",
				Slug:   "",
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					SectionTitles: []string{},
				},
			},
		},
		{
			name: "title-only-chapter",
			args: args{
				rawContent: `+++
title = "Prepare"
+++`,
			},
			want: Content{
				Title: "Prepare",
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					SectionTitles: []string{},
				},
			},
		},
		{
			name: "state-only-chapter",
			args: args{
				rawContent: `+++
state = "incomplete"
+++`,
			},
			want: Content{
				State: Incomplete,
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					SectionTitles: []string{},
				},
			},
		},
		{
			name: "complete-chapter-without-state",
			args: args{
				rawContent: `+++
archetype = "chapter"
title = "Prepare"
+++
Episodes
--------

- bar
`,
			},
			want: Content{
				Title: "Prepare",
				Body: &IndexBody{
					HasEpisodes: true,
					State:       Incomplete,
				},
			},
		},
		{
			name: "complete-chapter-with-state",
			args: args{
				rawContent: `+++
archetype = "chapter"
title = "Prepare"
state = "complete"
+++
Episodes
--------

- bar
`,
			},
			want: Content{
				Title: "Prepare",
				State: Complete,
				Body: &IndexBody{
					HasEpisodes: true,
					State:       Incomplete,
				},
			},
		},
		{
			name: "almost-complete-page",
			args: args{
				rawContent: `+++
title = "Prepare"
state = "complete"
+++
Summary
-------

- bar

Main Video
----------

Topics
------

- bar

Related Videos
--------------

- bar

Related Links
-------------

- bar

Exercises
---------

- bar
`,
			},
			want: Content{
				Title:  "Prepare",
				State:  Complete,
				Weight: "",
				Slug:   "",
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					HasSummary:      true,
					HasTopics:       true,
					HasExercises:    true,
					HasRelatedLinks: true,
					RelatedVideos:   Videos{},
					SectionTitles: []string{
						sectionSummary,
						sectionMainVideo,
						sectionTopics,
						sectionRelatedVideos,
						sectionRelatedLinks,
						sectionExercises,
					},
				},
			},
		},
		{
			name: "complete-page",
			args: args{
				rawContent: `+++
title = "Prepare"
state = "complete"
+++
Summary
-------

- bar

Main Video
----------

- bar

Topics
------

- bar

Related Videos
--------------

- bar

Related Links
-------------

- bar

Exercises
---------

- bar
`,
			},
			want: Content{
				Title:  "Prepare",
				State:  Complete,
				Weight: "",
				Slug:   "",
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					HasSummary:      true,
					HasTopics:       true,
					HasExercises:    true,
					HasRelatedLinks: true,
					RelatedVideos:   Videos{},
					SectionTitles: []string{
						sectionSummary,
						sectionMainVideo,
						sectionTopics,
						sectionRelatedVideos,
						sectionRelatedLinks,
						sectionExercises,
					},
				},
			},
		},
		{
			name: "incomplete-if-practice-is-missing",
			args: args{
				rawContent: `+++
title = "Prepare"
state = "complete"
+++
## Summary

- bar

## Main Video

- bar

## Topics

- bar

## Related Videos

- bar

## Related Links

- bar
`,
			},
			want: Content{
				Title:  "Prepare",
				State:  Complete,
				Weight: "",
				Slug:   "",
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					HasSummary:      true,
					HasTopics:       true,
					HasExercises:    false,
					HasRelatedLinks: true,
					RelatedVideos:   Videos{},
					SectionTitles: []string{
						sectionSummary,
						sectionMainVideo,
						sectionTopics,
						sectionRelatedVideos,
						sectionRelatedLinks,
					},
				},
			},
		},
		{
			name: "complete-page-with-hashmark-headers",
			args: args{
				rawContent: `+++
title = "Prepare"
state = "complete"
weight = 9
+++
## Summary

- bar

## Main Video

- bar

## Topics

- bar

## Related Videos

- bar

## Related Links

- bar

## Exercises

- bar
`,
			},
			want: Content{
				Title:  "Prepare",
				State:  Complete,
				Weight: "9",
				Slug:   "",
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					HasSummary:      true,
					HasTopics:       true,
					HasExercises:    true,
					HasRelatedLinks: true,
					RelatedVideos:   Videos{},
					SectionTitles: []string{
						sectionSummary,
						sectionMainVideo,
						sectionTopics,
						sectionRelatedVideos,
						sectionRelatedLinks,
						sectionExercises,
					},
				},
			},
		},
		{
			name: "bug unclear",
			args: args{
				rawContent: `+++
title = 'What Your Text Editor Says About You'
date = 2024-07-21T12:31:33+02:00
weight = 60
state = 'complete'
draft = false
slug = 'what-your-text-editor-says-about-you'
tags = ["no-exercise", "fun", "vim", "vscode", "goland", "jetbrains"]
disableMermaid = true
disableOpenapi = true
audience = 'all'
audienceImportance = 'optional'
+++

Main Video
----------

{{< time 5 >}}

This is just a fun video, don't take it too seriously. But also it's good to know what others will think about you based
on your choice of text editor. :D

{{< youtube sbdFwFDTDqU >}}
`,
			},
			want: Content{
				Title:  "What Your Text Editor Says About You",
				State:  Complete,
				Weight: "60",
				Slug:   "what-your-text-editor-says-about-you",
				Body: DefaultBody{
					Main: Main{
						Status: VideoPresent,
						Videos: []Video{
							{
								Badges:  Badges{},
								Issues:  nil,
								Minutes: 5,
								Valid:   true,
							},
						},
					},
					HasSummary:      false,
					HasTopics:       false,
					HasExercises:    true,
					HasRelatedLinks: false,
					RelatedVideos:   nil,
					SectionTitles: []string{
						sectionMainVideo,
					},
				},
				Audience:   All,
				Importance: Optional,
				Tags:       []string{"no-exercise", "fun", "vim", "vscode", "goland", "jetbrains"},
			},
		},
		{
			name: "practice",
			args: args{
				rawContent: `+++
title = 'Data Cleanup'
date = 2024-07-09T19:26:57+02:00
weight = 20
state = 'complete'
draft = false
slug = 'data-cleanup'
tags = ["vim", "practice"]
disableMermaid = true
disableOpenapi = true
audience = "all"
audienceImportance = "important"
+++

Description
-----------

Download [this SQL File](/a1.1/practice-data-cleanup.sql).

At this point, you don't really need to understand what this file is about, all you need to know is that we want to
turn it into a JSON file, using the values found in the second parentheses.

So basically your task is not turn that file into something that looks like this:

You don't need to worry about the white spaces the following two examples are also acceptable solutions:

### Examples

#### Example 1

#### Example 2

#### Example 3

### Hints

**Hint:** Arguably the fastest solution is using an editor with Vim motions and Vim macros, but other solutions are fine
as well. If you're familiar with tools like grep, sed or awk, those can be quite efficient for tasks like this too.

Recommended challenges
----------------------

### Display overall stats

Write an app to display the coordinates (x, y) for the largest, and smallest values for the whole dataset.

Example output:


### Display stats for each chart

Write an app to display the coordinates (x, y) for the largest, and smallest values for each chart.

So an example output could be the following:

Note that the order of the stat blocks does not matter, lemmy could come before lemmy.


Additional challenges
---------------------

{{<badge-extra>}}

### Sorting

This one is only different from the "Display chart stats" challenge is that here the order of the stats matter, they
should be ordered by the chart name, ordered Z to A, plus we should display all the coordinates where the value is the
maximum or minimum and make sure that they're ordered in incremental order.

Example output:

### Find the size of chart maps

Find a program that is able to tell the size of the map we have complete coordinates, meaning that no points are missing.

Example output:


### Find the size of intended chart maps and errors

People make mistakes, the provided SQL file is also imperfect. Modify your application so that it can give reviewers a
hint on missing coordinates or duplicates.

Example output:
`,
			},
			want: Content{
				Title:  "Data Cleanup",
				State:  Complete,
				Weight: "20",
				Slug:   "data-cleanup",
				Body: &PracticeBody{
					HasDescription:           true,
					HasRecommendedChallenges: true,
					HasAdditionalChallenges:  true,
				},
				Audience:   All,
				Importance: Important,
				Tags:       []string{"vim", "practice"},
			},
		},
		{
			name: "useful without video",
			args: args{
				rawContent: `+++
title = 'Free Dev Learning'
date = 2024-06-29T01:42:15+02:00
weight = 10
state = 'incomplete'
draft = false
slug = 'free-dev-learning'
tags = ["career", "learning", "no-exercise", "useful-without-video"]
disableMermaid = true
disableOpenapi = true
audience = "all"
audienceImportance = "relevant"
+++

Main Video
----------

{{<main-missing>}}

Related Links
-------------

### Platforms

- [exercism](https://exercism.org/)
`,
			},
			want: Content{
				Title:  "Free Dev Learning",
				State:  Incomplete,
				Weight: "10",
				Slug:   "free-dev-learning",
				Body: DefaultBody{
					Main: Main{
						Status: VideoMissing,
						Videos: nil,
					},
					HasSummary:         false,
					HasTopics:          false,
					HasExercises:       true,
					RelatedVideos:      nil,
					HasRelatedLinks:    true,
					UsefulWithoutVideo: true,
					SectionTitles:      []string{sectionMainVideo, sectionRelatedLinks},
				},
				Audience:   All,
				Importance: Relevant,
				Tags:       []string{"career", "learning", "no-exercise", "useful-without-video"},
			},
		},
		{
			name: "complex related videos section",
			args: args{
				rawContent: `+++
title = 'Electronic Computing'
date = 2024-07-28T11:34:54+02:00
weight = 80
state = 'complete'
draft = false
slug = 'electronic-computing'
tags = ["computer-science", "no-exercise"]
disableMermaid = true
disableOpenapi = true
audience = "all"
audienceImportance = "optional"
+++

Main Video
----------

{{< time 11 >}} {{<badge-audio>}} {{<badge-unchecked>}}

Thanks to the [Carrie Anne](https://about.me/carrieannephilbin) and [Crash Course](https://www.youtube.com/@crashcourse)
I will not have to make a video of this topic.

{{< youtube id=LN0ucKNX0hc start=56 >}}

Summary
-------

In this lesson, we will learn about the history of electronic computing, starting from the first programmable computer
to the first general-purpose computer.

Topics
------

- [Harvard Mark I](https://en.wikipedia.org/wiki/Harvard_Mark_I) - 1944, The first programmable computer
- [Mechanical relay](https://en.wikipedia.org/wiki/Relay) - An electrically operated switch
- [Grace Hopper](https://en.wikipedia.org/wiki/Grace_Hopper) - The inventor of the compiler
- [John Ambrose Fleming](https://en.wikipedia.org/wiki/John_Ambrose_Fleming) - The inventor of the vacuum tube
- [Vacuum tube](https://en.wikipedia.org/wiki/Vacuum_tube) - An electronic device used to control the flow of electric current
- [Diode](https://en.wikipedia.org/wiki/Diode) - A semiconductor device used to control the flow of electric current
- [Lee de Forest](https://en.wikipedia.org/wiki/Lee_de_Forest) - The inventor of the triode
- [Colossus](https://en.wikipedia.org/wiki/Colossus_computer) - 1943, The first programmable digital computer
- [Bill Tutte](https://en.wikipedia.org/wiki/Bill_Tutte) - The codebreaker who cracked the German Lorenz cipher
- [Tommy Flowers](https://en.wikipedia.org/wiki/Tommy_Flowers) - The engineer who built the Colossus
- [Alan Turing](https://en.wikipedia.org/wiki/Alan_Turing) - The father of computer science
- [The Bombe](https://en.wikipedia.org/wiki/Bombe) - 1940, A device used to break the German Enigma code
- [Enigma machine](https://en.wikipedia.org/wiki/Enigma_machine) - 1930, A German cipher machine used by soldiers
- [Lorenz](https://en.wikipedia.org/wiki/Lorenz_cipher) - A German cipher machine used by high-command
- [ENIAC](https://en.wikipedia.org/wiki/ENIAC) - 1945, The first general-purpose computer
- [John Mauchly](https://en.wikipedia.org/wiki/John_Mauchly) - The co-inventor of the ENIAC
- [J. Presper Eckert](https://en.wikipedia.org/wiki/J._Presper_Eckert) - The co-inventor of the ENIAC
- [Transistor](https://en.wikipedia.org/wiki/Transistor) - A semiconductor device used to amplify or switch electronic signals
- [William Shockley](https://en.wikipedia.org/wiki/William_Shockley) - The co-inventor of the transistor
- [John Bardeen](https://en.wikipedia.org/wiki/John_Bardeen) - The co-inventor of the transistor
- [Walter Brattain](https://en.wikipedia.org/wiki/Walter_Houser_Brattain) - The co-inventor of the transistor
- [Gate electrode](https://en.wikipedia.org/wiki/Gate_(transistor)) - The electrode that controls the flow of electric current
- [Silicon Valley](https://en.wikipedia.org/wiki/Silicon_Valley) - The center of the computer industry

Related Videos
--------------

### The Analytical Engine (Charles Babbage, Ada Lovelace)

#### The greatest machine that never was - John Graham-Cumming - TED-Ed

{{< time 12 >}} {{<badge-extra>}}

{{< youtube FlfChYGv3Z4 >}}

#### Babbage's Analytical Engine - Computerphile

{{< time 14 >}} {{<badge-extra>}}

{{< youtube 5rtKoKFGFSM >}}

#### Ada Lovelace: The First Computer Programmer - Biographics

{{< time 21 >}} {{<badge-extra>}}

{{< youtube id=IZptxisyVqQ start=60 >}}

---

### Harvard Mark I

#### Supercomputer Where It All Started - Harvard Mark 1 - Major Hardware

{{< time 6 >}} {{<badge-extra>}}

{{< youtube cd2DV-AoCk4 >}}

#### Harvard Mark I, 2022 - CS50

{{< time 3 >}} {{<badge-extra>}}

{{< youtube 7l8W96I7_ew >}}

---

### Enigma, Bombe (Alan Turing)

#### How did the Enigma Machine work? - Jared Owen

{{< time 20 >}} {{<badge-extra>}}

{{< youtube ybkkiGtJmkM >}}


### Lorenz and Colossus (Tommy Flowers, Bill Tutte)

#### Why the Toughest Code to Break in WW2 WASN'T Enigma - The Story of the Lorenz Cipher

{{< time 11 >}} {{<badge-extra>}}

{{< youtube RCWgOaDOzpY >}}

#### Colossus & Bletchley Park - Computerphile

{{< time 9 >}} {{<badge-extra>}}

{{< youtube 9HH-asvLAj4 >}}

#### Colossus - The Greatest Secret in the History of Computing - The Centre for Computing History

{{< time 60 >}} {{<badge-extra>}}

This is not only about Colossus, but provides a lot of context, including basic cryptographic problems of the time. It's
probably my favorite video recommended on this page.

{{< youtube g2tMcMQqSbA >}}

### Why Build Colossus? (Bill Tutte) - Computerphile

{{< time 8 >}} {{<badge-extra>}}

{{< youtube 1f82-aTYNb8 >}}

---

### Transistors and ENIAC (John Mauchly, J. Presper Eckert)

#### Transistors - The Invention That Changed The World - Real Engineering

{{< time 8 >}} {{<badge-extra>}}

{{< youtube OwS9aTE2Go4 >}}
`,
			},
			want: Content{
				Title:  "Electronic Computing",
				State:  Complete,
				Weight: "80",
				Slug:   "electronic-computing",
				Body: DefaultBody{
					Main: Main{
						Status: VideoPresent,
						Videos: []Video{
							{
								Badges:  Badges{Unchecked},
								Issues:  nil,
								Minutes: 11,
								Valid:   true,
							},
						},
					},
					HasSummary:   true,
					HasTopics:    true,
					HasExercises: true,
					RelatedVideos: Videos{
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 12,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 14,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 21,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 6,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 3,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 20,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 11,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 9,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 60,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 8,
							Valid:   true,
						},
						{
							Badges:  Badges{Extra},
							Issues:  nil,
							Minutes: 8,
							Valid:   true,
						},
					},
					HasRelatedLinks:    false,
					UsefulWithoutVideo: false,
					SectionTitles: []string{
						sectionMainVideo,
						sectionSummary,
						sectionTopics,
						sectionRelatedVideos,
					},
				},
				Audience:   All,
				Importance: Optional,
				Tags:       []string{"computer-science", "no-exercise"},
			},
		},
		{
			name: "main-in-wrong-order",
			args: args{
				rawContent: `+++
title = 'Advanced Linux Commands'
date = 2024-06-28T21:39:29+02:00
weight = 40
state = 'incomplete'
draft = false
slug = 'advanced-linux-commands'
tags = ["linux", "cli"]
disableMermaid = true
disableOpenapi = true
audience = "all"
audienceImportance = "important"
+++

Summary
-------

Topics
------

- [which](https://linux.die.net/man/1/which)
- [ping](https://linux.die.net/man/1/ping)
- [whoami](https://linux.die.net/man/1/whoami)
- [whatis](https://linux.die.net/man/1/whatis)
- [find](https://linux.die.net/man/1/find)
- [uname](https://linux.die.net/man/1/uname)
- [lsb_release](https://linux.die.net/man/1/lsb_release)
- [curl](https://linux.die.net/man/1/curl)
- [wget](https://linux.die.net/man/1/wget)
- [httping](https://linux.die.net/man/1/httping)
- [alias](https://linux.die.net/man/1/alias)
- [unalias](https://linux.die.net/man/1/unalias)
- [wc](https://linux.die.net/man/1/wc)
- [cut](https://linux.die.net/man/1/cut)
- [awk](https://linux.die.net/man/1/awk)
- [sed](https://linux.die.net/man/1/sed)
- [pgrep](https://linux.die.net/man/1/pgrep)
- [fuser](https://linux.die.net/man/1/fuser)
- [rsync](https://linux.die.net/man/1/rsync)
- [tar](https://linux.die.net/man/1/tar)
- [gzip](https://linux.die.net/man/1/gzip)
- [lsof](https://linux.die.net/man/1/lsof)
- [screen](https://linux.die.net/man/1/screen)

Main Video
----------

Related Videos
--------------

### Linux Command Line for Beginners

{{< time 59 >}} {{<badge-alternative>}}

{{< youtube 16d2lHc0Pe8 >}}

### 50 MUST KNOW Linux Commands (in under 15 minutes)

{{< time 14 >}} {{<badge-alternative>}}

{{< youtube nzjkbQNmXAE >}}

Exercises
---------

`,
			},
			want: Content{
				Title:  "Advanced Linux Commands",
				State:  Incomplete,
				Weight: "40",
				Slug:   "advanced-linux-commands",
				Body: DefaultBody{
					Main: Main{
						Status: VideoProblem,
						Videos: nil,
					},
					HasSummary:   false,
					HasTopics:    true,
					HasExercises: false,
					RelatedVideos: Videos{
						{
							Badges:  Badges{Alternative},
							Minutes: 59,
							Valid:   true,
						},
						{
							Badges:  Badges{Alternative},
							Minutes: 14,
							Valid:   true,
						},
					},
					HasRelatedLinks: false,
					SectionTitles: []string{
						sectionSummary,
						sectionTopics,
						sectionMainVideo,
						sectionRelatedVideos,
						sectionExercises,
					},
					UsefulWithoutVideo: false,
				},
				Audience:   All,
				Importance: Important,
				Tags:       []string{"linux", "cli"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// execute
			got, err := ParseMarkdown(tt.args.rawContent)
			require.NoError(t, err)

			// verify
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractRelatedVideos(t *testing.T) {
	type args struct {
		content     string
		noBadgeOkay bool
	}
	tests := []struct {
		name string
		args args
		want Videos
	}{
		{
			name: "empty content is skipped",
			args: args{
				content:     "",
				noBadgeOkay: false,
			},
			want: nil,
		},
		{
			name: "no badges lead to issues",
			args: args{
				content: `### This is a title\n\nfoo\n
{{< time 5 >}}

{{< youtube abc >}}`,
				noBadgeOkay: false,
			},
			want: Videos{
				{
					Badges: Badges{},
					Issues: []string{
						"missing badge shortcode",
					},
					Minutes: 5,
					Valid:   true,
				},
			},
		},
		{
			name: "too many badges lead to issues",
			args: args{
				content: `### This is it
{{< time 5 >}} {{<time  12>}}

{{<  badge-extra   >}} {{<badge-extra>}}

{{< youtube abc >}} {{<youtube def>}}
`,
				noBadgeOkay: false,
			},
			want: Videos{
				{
					Badges: Badges{Extra, Extra},
					Issues: []string{
						"multiple time shortcodes found",
						"unexpected badge shortcode found: extra",
						"multiple youtube shortcodes found",
					},
					Minutes: 5,
					Valid:   true,
				},
			},
		},
		{
			name: "multiple badges, multiple issues",
			args: args{
				content: `### Almost empty

foo

### Missing badge

{{< time 5 >}}

{{< youtube abc >}}

### Multiple badge

{{< time 123 >}} {{<badge-alternative>}} {{<badge-extra>}}

{{< youtube foo >}}

### Multiple youtube videos

{{< time 17 >}} {{<badge-extra>}}

{{< youtube bar >}}
{{< youtube foo >}}
`,
				noBadgeOkay: false,
			},
			want: Videos{
				{
					Badges: Badges{},
					Issues: []string{
						"missing badge shortcode",
					},
					Minutes: 5,
					Valid:   true,
				},
				{
					Badges: Badges{Alternative, Extra},
					Issues: []string{
						"unexpected badge shortcode found: extra",
						"badges should have full-course, but do not. badges: alternative, extra",
					},
					Minutes: 123,
					Valid:   true,
				},
				{
					Badges: Badges{Extra},
					Issues: []string{
						"multiple youtube shortcodes found",
					},
					Minutes: 17,
					Valid:   true,
				},
			},
		},
		{
			name: "success",
			args: args{
				content: `### Multiple youtube videos

{{< time 17 >}} {{<badge-extra>}} {{<badge-unchecked>}} {{<badge-audio>}}

{{< youtube bar >}}
`,
				noBadgeOkay: false,
			},
			want: Videos{
				{
					Badges:  Badges{Extra, Unchecked},
					Issues:  nil,
					Minutes: 17,
					Valid:   true,
				},
			},
		},
		{
			name: "success - empty skipped",
			args: args{
				content: `### Skipped
### Multiple youtube videos

{{< time 17 >}} {{<badge-extra>}}

{{< youtube bar >}}
`,
				noBadgeOkay: false,
			},
			want: Videos{
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 17,
					Valid:   true,
				},
			},
		},
		{
			name: "success - no-embed badge does not count as badge",
			args: args{
				content: `### Multiple youtube videos

{{< time 17 >}} {{<badge-extra>}} {{<badge-no-embed>}}
`,
				noBadgeOkay: false,
			},
			want: Videos{
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 17,
					Valid:   true,
				},
			},
		},
		{
			name: "failure - no-embed badge does count as a video",
			args: args{
				content: `### Multiple youtube videos

{{< time 17 >}} {{<badge-extra>}} {{<badge-no-embed>}}

{{< youtube bar >}}
`,
				noBadgeOkay: false,
			},
			want: Videos{
				{
					Badges:  Badges{Extra},
					Issues:  []string{"unexpected youtube shortcode together with no-embed badge"},
					Minutes: 17,
					Valid:   true,
				},
			},
		},
		{
			name: "complex video section",
			args: args{
				content: `### The Analytical Engine (Charles Babbage, Ada Lovelace)

#### The greatest machine that never was - John Graham-Cumming - TED-Ed

{{< time 12 >}} {{<badge-extra>}}

{{< youtube FlfChYGv3Z4 >}}

#### Babbage's Analytical Engine - Computerphile

{{< time 14 >}} {{<badge-extra>}}

{{< youtube 5rtKoKFGFSM >}}

#### Ada Lovelace: The First Computer Programmer - Biographics

{{< time 21 >}} {{<badge-extra>}}

{{< youtube id=IZptxisyVqQ start=60 >}}

---

### Harvard Mark I

#### Supercomputer Where It All Started - Harvard Mark 1 - Major Hardware

{{< time 6 >}} {{<badge-extra>}}

{{< youtube cd2DV-AoCk4 >}}

#### Harvard Mark I, 2022 - CS50

{{< time 3 >}} {{<badge-extra>}}

{{< youtube 7l8W96I7_ew >}}

---

### Enigma, Bombe (Alan Turing)

#### How did the Enigma Machine work? - Jared Owen

{{< time 20 >}} {{<badge-extra>}}

{{< youtube ybkkiGtJmkM >}}


### Lorenz and Colossus (Tommy Flowers, Bill Tutte)

#### Why the Toughest Code to Break in WW2 WASN'T Enigma - The Story of the Lorenz Cipher

{{< time 11 >}} {{<badge-extra>}}

{{< youtube RCWgOaDOzpY >}}

#### Colossus & Bletchley Park - Computerphile

{{< time 9 >}} {{<badge-extra>}}

{{< youtube 9HH-asvLAj4 >}}

#### Colossus - The Greatest Secret in the History of Computing - The Centre for Computing History

{{< time 60 >}} {{<badge-extra>}}

This is not only about Colossus, but provides a lot of context, including basic cryptographic problems of the time. It's
probably my favorite video recommended on this page.

{{< youtube g2tMcMQqSbA >}}

### Why Build Colossus? (Bill Tutte) - Computerphile

{{< time 8 >}} {{<badge-extra>}}

{{< youtube 1f82-aTYNb8 >}}


---

### Transistors and ENIAC (John Mauchly, J. Presper Eckert)

#### Transistors - The Invention That Changed The World - Real Engineering

{{< time 8 >}} {{<badge-extra>}}

{{< youtube OwS9aTE2Go4 >}}
`,
				noBadgeOkay: false,
			},
			want: Videos{
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 12,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 14,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 21,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 6,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 3,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 20,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 11,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 9,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 60,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 8,
					Valid:   true,
				},
				{
					Badges:  Badges{Extra},
					Issues:  nil,
					Minutes: 8,
					Valid:   true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// execute
			got := ExtractVideos(tt.args.content, tt.args.noBadgeOkay)

			// verify
			assert.Equal(t, tt.want, got)
		})
	}
}
