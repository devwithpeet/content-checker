package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCourses_Add(t *testing.T) {
	stubContent := Content{
		Title: "Baz",
		State: Complete,
		Body: DefaultBody{
			Main: Main{
				Status: VideoProblem,
				Videos: nil,
			},
			HasExercises: true,
		},
	}

	type args struct {
		filePath  string
		courseFN  string
		chapterFN string
		pageFN    string
		content   Content
	}
	tests := []struct {
		name string
		args args
		c    Courses
		want Courses
	}{
		{
			name: "add to empty",
			args: args{
				filePath:  "foo/bar/baz.md",
				courseFN:  "foo",
				chapterFN: "bar",
				pageFN:    "baz.md",
				content:   stubContent,
			},
			c: Courses{},
			want: Courses{
				{
					Title: "foo",
					Chapters: Chapters{
						{
							Title: "bar",
							Pages: Pages{
								{
									FilePath: "foo/bar/baz.md",
									Title:    "baz.md",
									Content:  stubContent,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "add to pages",
			args: args{
				filePath:  "foo/bar/baz.md",
				courseFN:  "foo",
				chapterFN: "bar",
				pageFN:    "baz.md",
				content:   stubContent,
			},
			c: Courses{
				{
					Title: "foo",
					Chapters: Chapters{
						{
							Title: "bar",
							Pages: Pages{
								{
									FilePath: "baz0.md",
									Content:  stubContent,
								},
							},
						},
					},
				},
			},
			want: Courses{
				{
					Title: "foo",
					Chapters: Chapters{
						{
							Title: "bar",
							Pages: Pages{
								{
									FilePath: "baz0.md",
									Content:  stubContent,
								},
								{
									FilePath: "foo/bar/baz.md",
									Title:    "baz.md",
									Content:  stubContent,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "add to chapters",
			args: args{
				filePath:  "foo/bar/baz.md",
				courseFN:  "foo",
				chapterFN: "bar",
				pageFN:    "baz.md",
				content:   stubContent,
			},
			c: Courses{
				{
					Title: "foo",
					Chapters: Chapters{
						{
							Title: "bar0",
							Pages: Pages{
								{
									FilePath: "baz.md",
									Content:  stubContent,
								},
							},
						},
					},
				},
			},
			want: Courses{
				{
					Title: "foo",
					Chapters: Chapters{
						{
							Title: "bar0",
							Pages: Pages{
								{
									FilePath: "baz.md",
									Content:  stubContent,
								},
							},
						},
						{
							Title: "bar",
							Pages: Pages{
								{
									FilePath: "foo/bar/baz.md",
									Title:    "baz.md",
									Content:  stubContent,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "add to courses",
			args: args{
				filePath:  "foo/bar/baz.md",
				courseFN:  "foo",
				chapterFN: "bar",
				pageFN:    "baz.md",
				content:   stubContent,
			},
			c: Courses{
				{
					Title: "foo0",
					Chapters: Chapters{
						{
							Title: "bar",
							Pages: Pages{
								{
									FilePath: "baz.md",
									Content:  stubContent,
								},
							},
						},
					},
				},
			},
			want: Courses{
				{
					Title: "foo0",
					Chapters: Chapters{
						{
							Title: "bar",
							Pages: Pages{
								{
									FilePath: "baz.md",
									Content:  stubContent,
								},
							},
						},
					},
				},
				{
					Title: "foo",
					Chapters: Chapters{
						{
							Title: "bar",
							Pages: Pages{
								{
									FilePath: "foo/bar/baz.md",
									Title:    "baz.md",
									Content:  stubContent,
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// execute
			got := tt.c.Add(tt.args.filePath, tt.args.courseFN, tt.args.chapterFN, tt.args.pageFN, tt.args.content)

			// verify
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isOrderedCorrectly(t *testing.T) {
	type args struct {
		goldenMap  map[string]int
		givenSlice []string
	}
	tests := []struct {
		name             string
		args             args
		wantFirstFailure string
		wantOK           bool
	}{
		{
			name: "empty",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{},
			},
			wantFirstFailure: "",
			wantOK:           true,
		},
		{
			name: "main-only",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{sectionMainVideo},
			},
			wantFirstFailure: "",
			wantOK:           true,
		},
		{
			name: "notes-only",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{sectionNotes},
			},
			wantFirstFailure: "",
			wantOK:           true,
		},
		{
			name: "main-notes",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{sectionMainVideo, sectionNotes},
			},
			wantFirstFailure: "",
			wantOK:           true,
		},
		{
			name: "notes-main",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{sectionNotes, sectionMainVideo},
			},
			wantFirstFailure: sectionMainVideo,
			wantOK:           false,
		},
		{
			name: "notes-notes",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{sectionNotes, sectionNotes},
			},
			wantFirstFailure: sectionNotes,
			wantOK:           false,
		},
		{
			name: "main-main",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{sectionMainVideo, sectionMainVideo},
			},
			wantFirstFailure: sectionMainVideo,
			wantOK:           false,
		},
		{
			name: "main-notes-main",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{sectionMainVideo, sectionNotes, sectionMainVideo},
			},
			wantFirstFailure: sectionMainVideo,
			wantOK:           false,
		},
		{
			name: "unexpected",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{"unexpected"},
			},
			wantFirstFailure: "unexpected",
			wantOK:           false,
		},
		{
			name: "unexpected after main",
			args: args{
				goldenMap:  defaultBodySectionMap,
				givenSlice: []string{sectionMainVideo, "unexpected"},
			},
			wantFirstFailure: "unexpected",
			wantOK:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// execute
			gotFirstFailure, gotOK := isOrderedCorrectly(tt.args.goldenMap, tt.args.givenSlice)

			// verify
			assert.Equalf(t, tt.wantFirstFailure, gotFirstFailure, "isOrderedCorrectly(%v, %v)", tt.args.goldenMap, tt.args.givenSlice)
			assert.Equalf(t, tt.wantOK, gotOK, "isOrderedCorrectly(%v, %v)", tt.args.goldenMap, tt.args.givenSlice)
		})
	}
}

func Test_slugify(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "complex",
			args: args{title: "CQRS - Command Query Responsibility Segregation?"},
			want: "cqrs-command-query-responsibility-segregation",
		},
		{
			name: "c#",
			args: args{title: "C# Basics"},
			want: "c-sharp-basics",
		},
		{
			name: ".net",
			args: args{title: "About the .NET Framework?"},
			want: "about-the-dot-net-framework",
		},
		{
			name: "web i.",
			args: args{title: "Web I."},
			want: "web-i",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// execute
			got := slugify(tt.args.title)

			// verify
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_CalculateStatus(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "complex",
			args: args{title: "CQRS - Command Query Responsibility Segregation?"},
			want: "cqrs-command-query-responsibility-segregation",
		},
		{
			name: "c#",
			args: args{title: "C# Basics"},
			want: "c-sharp-basics",
		},
		{
			name: ".net",
			args: args{title: "About the .NET Framework?"},
			want: "about-the-dot-net-framework",
		},
		{
			name: "web i.",
			args: args{title: "Web I."},
			want: "web-i",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// execute
			got := slugify(tt.args.title)

			// verify
			assert.Equal(t, tt.want, got)
		})
	}
}
