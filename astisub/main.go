package main

import (
	"flag"
	"fmt"
	"github.com/asticode/go-astisub"
	"log"
	"strconv"
	"strings"

	"github.com/asticode/go-astikit"
)

// Flags
var (
	actual1          = flag.Duration("a1", 0, "the first actual duration")
	actual2          = flag.Duration("a2", 0, "the second actual duration")
	desired1         = flag.Duration("d1", 0, "the first desired duration")
	desired2         = flag.Duration("d2", 0, "the second desired duration")
	fragmentDuration = flag.Duration("f", 0, "the fragment duration")
	inputPath        = astikit.NewFlagStrings()
	teletextPage     = flag.Int("p", 0, "the teletext page")
	outputPath       = flag.String("o", "", "the output path")
	syncDuration     = flag.Duration("s", 0, "the sync duration")
	segmentationType = flag.String("st", "UNIFIED", "the segmentation type UNIFIED/SPECIFIED. In unified"+
		" segmentation all segment have same duration whereas specified have user given duration for each segment.")
	segmentDuration  = flag.Float64("sd", 5, "segmentation duration for unified segmentation type")
	segmentDurations = flag.String("sds", "", "segment durations for all segments seperated by comma")
	webvttOffset     = flag.Float64("wo", 0, "webvtt offset for synchronization of segment in hls")
	timecodeOffset   = flag.Float64("to", 0, "timecode offset to modify start timecode")
)

func main() {
	// Init
	cmd := astikit.FlagCmd()
	flag.Var(&inputPath, "i", "the input paths")
	flag.Parse()

	// Validate input path
	if len(*inputPath.Slice) == 0 {
		log.Fatal("Use -i to provide at least one input path")
	}

	// Validate output path
	if len(*outputPath) <= 0 {
		log.Fatal("Use -o to provide an output path")
	}

	// Open first input path
	var sub *astisub.Subtitles
	var err error
	if sub, err = astisub.Open(astisub.Options{Filename: (*inputPath.Slice)[0], Teletext: astisub.TeletextOptions{Page: *teletextPage}}); err != nil {
		log.Fatalf("%s while opening %s", err, (*inputPath.Slice)[0])
	}

	// Switch on subcommand
	switch cmd {
	case "apply-linear-correction":
		// Validate actual and desired durations
		if *actual1 <= 0 {
			log.Fatal("Use -a1 to provide the first actual duration")
		}
		if *desired1 <= 0 {
			log.Fatal("Use -d1 to provide the first desired duration")
		}
		if *actual2 <= 0 {
			log.Fatal("Use -a2 to provide the second actual duration")
		}
		if *desired2 <= 0 {
			log.Fatal("Use -d2 to provide the second desired duration")
		}

		// Apply linear correction
		sub.ApplyLinearCorrection(*actual1, *desired1, *actual2, *desired2)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "convert":
		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "fragment":
		// Validate fragment duration
		if *fragmentDuration <= 0 {
			log.Fatal("Use -f to provide a fragment duration")
		}

		// Fragment
		sub.Fragment(*fragmentDuration)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "merge":
		// Validate second input path
		if len(*inputPath.Slice) == 1 {
			log.Fatal("Use -i to provide at least two input paths")
		}

		// Open second input path
		var sub2 *astisub.Subtitles
		if sub2, err = astisub.Open(astisub.Options{Filename: (*inputPath.Slice)[1], Teletext: astisub.TeletextOptions{Page: *teletextPage}}); err != nil {
			log.Fatalf("%s while opening %s", err, (*inputPath.Slice)[1])
		}

		// Merge
		sub.Merge(sub2)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "optimize":
		// Optimize
		sub.Optimize()

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "sync":
		// Validate sync duration
		if *syncDuration == 0 {
			log.Fatal("Use -s to provide a sync duration")
		}

		// Fragment
		sub.Add(*syncDuration)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "unfragment":
		// Unfragment
		sub.Unfragment()

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "vtt-segment":
		// go run main.go vtt-segment -i <input_path>  -st SPECIFIED -sds "2,4,10,12,13" -o "seg-%03d.vtt" -wo 10
		if *segmentationType == "SPECIFIED" && segmentDurations == nil {
			log.Fatalf("segment duration must be present for specified segmentation type")
		}
		var sds []float64
		if *segmentDurations != "" {
			for _, segDuration := range strings.Split(*segmentDurations, ",") {
				segDur, err := strconv.ParseFloat(segDuration, 64)
				if err != nil {
					log.Fatalf("%s unable to parse seg duration %s", err, segDuration)
				}
				sds = append(sds, segDur)
			}
		}
		segmentedSubs := sub.Segment(*segmentationType, *segmentDuration, sds)
		for idx, segmentedSub := range segmentedSubs {
			if err = segmentedSub.WriteToWebVTTFile(fmt.Sprintf(*outputPath, idx), *webvttOffset); err != nil {
				log.Fatalf("%s while writing to %s", err, *outputPath)
			}
		}
	case "modify-start-timecode":
		if err = sub.ModifyStartTimeCode(*timecodeOffset); err != nil {
			log.Fatalf("%s while modify start time code to %s", err, *outputPath)
		}
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	default:
		log.Fatalf("Invalid subcommand %s", cmd)
	}
}
