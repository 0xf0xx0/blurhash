package main

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/bbrks/go-blurhash"
	"github.com/kovidgoyal/imaging"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:                   "blurhash",
		Usage:                  "scales down and blurhashes images, preserving aspect ratio",
		UsageText:              "blurhash [options] <image file>",
		Version:                "0.1.0",
		UseShortOptionHandling: true,
		EnableShellCompletion:  true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Value:   "-",
				Aliases: []string{"o"},
				Usage:   "File to output to (can be - for stdout)",
			},
			&cli.UintFlag{
				Name:    "xcomponent",
				Value:   3,
				Aliases: []string{"x"},
				Usage:   "number of X components",
				Action: func(_ context.Context, _ *cli.Command, v uint) error {
					if v < 1 || v > 9 {
						return fmt.Errorf("invalid X component %d (range: 1-9)", v)
					}
					return nil
				},
			},
			&cli.UintFlag{
				Name:    "ycomponent",
				Value:   4,
				Aliases: []string{"y"},
				Usage:   "number of Y components",
				Action: func(_ context.Context, _ *cli.Command, v uint) error {
					if v < 1 || v > 9 {
						return fmt.Errorf("invalid Y component %d (range: 1-9)", v)
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:    "decode",
				Value:   false,
				Aliases: []string{"d"},
				Usage:   "decode a blurhash string",
			},
			&cli.UintFlag{
				Name:    "width",
				Value:   3,
				Aliases: []string{"w"},
				Usage:   "width (used with --decode)",
				Action: func(_ context.Context, _ *cli.Command, v uint) error {
					return nil
				},
			},
			&cli.UintFlag{
				Name:    "height",
				Value:   32,
				Aliases: []string{"h"},
				Usage:   "height (used with --decode)",
				Action: func(_ context.Context, _ *cli.Command, v uint) error {
					return nil
				},
			},
		},
		Action: func(_ context.Context, ctx *cli.Command) error {
			input := resolvePath(ctx.Args().First())
			output := resolvePath(ctx.String("output"))
			xcomp := int(ctx.Uint("xcomponent"))
			ycomp := int(ctx.Uint("ycomponent"))
			decode := ctx.Bool("decode")

			if input == "" {
				return cli.Exit("no input given", 1)
			}

			if decode {
				width := int(ctx.Uint("width"))
				height := int(ctx.Uint("height"))
				img, err := blurhash.Decode(input, width, height, 1)
				if err != nil {
					return cli.Exit(fmt.Sprintf("Error during decoding: %s", err), 2)
				}

				if output == "-" {
					png.Encode(os.Stdout, img)
				} else {
					file, err := os.Create(output)
					if err != nil {
						return cli.Exit(fmt.Sprintf("Error during writing: %s", err), 2)
					}
					png.Encode(file, img)
				}

				return cli.Exit("", 0)
			}

			/// encode

			file, err := os.Open(resolvePath(input))
			if err != nil {
				return cli.Exit(fmt.Sprintf("Input could not be opened: %s", err), 1)
			}
			defer file.Close()

			rawImg, _, err := image.Decode(file)
			if err != nil {
				println(fmt.Sprintf("Input could not be decoded: %s", err))
				return cli.Exit(fmt.Sprintf("Input could not be decoded: %s", err), 1)
			}
			sw, sh := scaleDown(rawImg.Bounds().Dx(), rawImg.Bounds().Dy())
			resized := imaging.Resize(rawImg, sw, sh, imaging.BSpline)
			hash, err := blurhash.Encode(xcomp, ycomp, resized)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error during encoding: %s", err), 2)
			}
			out := fmt.Sprintf("%s %d,%d", hash, sw, sh)
			if output == "-" {
				fmt.Println(out)
			} else {
				file, err := os.Create(output)
				if err != nil {
					return cli.Exit(fmt.Sprintf("Error during writing: %s", err), 2)
				}
				defer file.Close()
				fmt.Fprint(file, out)
			}
			return cli.Exit("", 0)
		},
	}
	app.Run(context.TODO(), os.Args)
}

// resolves ~ and cleans path
// https://stackoverflow.com/a/17617721
func resolvePath(path string) string {
	if strings.HasPrefix(path, "~") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		usr, _ := user.Current()
		home := usr.HomeDir
		path = filepath.Join(home, path[1:])
	}
	return path
}
func scaleDown(x, y int) (int, int) {
	divisor := float64(max(x, y)) / 32

	return int(float64(x) / divisor), int(float64(y) / divisor)
}
