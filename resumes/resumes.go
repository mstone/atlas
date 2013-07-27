package resumes

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

func L(s string, v ...interface{}) {
	if glog.V(1) {
		glog.Infof("resume "+s, v...)
	}
}

func handlePdf(src, fn, dst string) error {
	L("handlePdf(): src: %q, fn: %q, dst: %q", src, fn, dst)
	cmd := exec.Command("cp", "-f", src, dst)
	L("handlePdf: %s", cmd)
	return cmd.Run()
}

func runLibreOffice(src, dst string) error {
	dstDir := path.Dir(dst)
	tmpDir, err := ioutil.TempDir(dstDir, "atlas")
	if err != nil {
		L("runLibreOffice: warning: %q", err)
		return err
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("libreoffice",
		"--headless",
		"--convert-to",
		"pdf:writer_pdf_Export",
		"--outdir",
		tmpDir,
		src)

	L("runLibreOffice: %s", cmd)
	err = cmd.Run()
	if err != nil {
		L("runLibreOffice: warning: %q", err)
		return err
	}

	matches, err := filepath.Glob(path.Join(tmpDir, "*.pdf"))
	if err != nil {
		L("runLibreOffice: warning: %q", err)
		return err
	}
	if len(matches) != 1 {
		L("runLibreOffice: warning: len(%q) != 1", matches)
		return err
	}
	L("runLibreOffice: rename %q -> %q", matches[0], dst)
	err = os.Rename(matches[0], dst)

	return err
}

func runGhostScriptClip(dst string) error {
	cmd := exec.Command("gs",
		"-sDEVICE=bbox",
		"-dNOPAUSE",
		"-dSAFER",
		"-dBATCH",
		"-f",
		dst)
	L("runGhostScriptClip: %s", cmd)
	return cmd.Run()
}

func handleDoc(src, fn, dst string) error {
	L("handleDoc(): src: %q, fn: %q, dst: %q", src, fn, dst)
	return runLibreOffice(src, dst)
}

func handleDocx(src, fn, dst string) error {
	L("handleDocx(): src: %q, fn: %q, dst: %q", src, fn, dst)
	return runLibreOffice(src, dst)
}

func runPdftkBurst(src, pagesDir string) error {
	cmd := exec.Command("pdftk",
		src,
		"burst",
		"output",
		path.Join(pagesDir, "%02d.pdf"))
	L("runPdftkBurst: %s", cmd)
	return cmd.Run()
}

func runInkscape(src, dst string) error {
	cmd := exec.Command("inkscape",
		"-l",
		dst,
		src)
	L("runInkscape: %s", cmd)
	return cmd.Run()
}

func convertPdfPages(fn string, chart string, chartFile *os.File, svgPagesDir string, walkPath string, info os.FileInfo, walkErr error) error {
	L("convertPdfPages: fn: %q, walkPath: %q, walkErr: %q ", fn, walkPath, walkErr)
	if walkErr != nil {
		return walkErr
	}

	ext := filepath.Ext(walkPath)
	if ext == "" {
		return nil
	}

	base := filepath.Base(walkPath)

	pageFn := base[:len(base)-len(ext)]

	dst := path.Join(svgPagesDir, pageFn+".svg")

	err := runInkscape(walkPath, dst)
	if err != nil {
		L("convertPdfPages: inkscape warning: %s", err)
	}

	origSvgDst := path.Join(svgPagesDir, pageFn+".orig.svg")
	cmd := exec.Command("cp", "-f", dst, origSvgDst)
	err = cmd.Run()
	if err != nil {
		L("convertPdfPages: cp warning: %s", err)
	}

	origSvg := path.Join("svg_pages", fn, pageFn+".orig.svg")
	pageSvg := path.Join("svg_pages", fn, pageFn+".svg")
	pageEditor := path.Join(pageSvg, "editor")

	fmt.Fprintf(chartFile, "([edit page %s](%s), [see original copy](%s))\n", pageFn, pageEditor, origSvg)
	fmt.Fprintf(chartFile, "![](%s)\n", pageSvg)
	chartFile.WriteString("\n")

	return nil
}

func convert(inputPath, safeName, outputPath, displayName string) error {
	L("convert: inputPath: %q, safeName: %q, outputPath: %q, displayName", inputPath, safeName, outputPath, displayName)

	safeExt := filepath.Ext(safeName)
	if safeExt == "" {
		L("convert: safeName: %q has no extension; skipping.", safeName)
		return nil
	}

	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		L("convert: warning: unable to make chart dir %q, err: %q; skipping", outputPath, err)
		return err
	}

	chart := outputPath

	chartPath := path.Join(chart, "index.txt")

	_, err = os.Stat(chartPath)
	if err == nil {
		L("convert: warning: skipping input %q since chart %q already exists.", safeName, chartPath)
		return nil
	}

	chartFile, err := os.Create(chartPath)
	if err != nil {
		L("convert: warning:  unable to create chartPath: %q", chartPath)
		return nil
	}
	defer chartFile.Close()

	dstPdf := path.Join(chart, "input.pdf")

	_, err = os.Stat(dstPdf)
	if err != nil {
		if os.IsNotExist(err) {
			switch safeExt {
			default:
				L("walkInput: skipping %s", inputPath)
			case ".pdf":
				err = handlePdf(inputPath, safeName, dstPdf)
			case ".doc":
				err = handleDoc(inputPath, safeName, dstPdf)
			case ".docx":
				err = handleDocx(inputPath, safeName, dstPdf)
			}
			if err != nil {
				L("walkInput: ingest warning: %s", err)
			}
		} else {
			L("walkInput: stat warning: %s", err)
		}
	}

	err = runGhostScriptClip(dstPdf)
	if err != nil {
		L("walkInput: gs warning: %s", err)
	}

	pdfPagesDir := path.Join(chart, "pdf_pages", safeName)
	os.MkdirAll(pdfPagesDir, 0755)

	svgPagesDir := path.Join(chart, "svg_pages", safeName)
	os.MkdirAll(svgPagesDir, 0755)

	err = runPdftkBurst(dstPdf, pdfPagesDir)
	if err != nil {
		L("walkInput: pdftk warning: %s", err)
	}

	now := time.Now()
	date := fmt.Sprintf("%s %0.2d, %d", now.Month().String(), now.Day(), now.Year())

	chartFile.WriteString("% Resume: " + displayName + "\n")
	chartFile.WriteString("% Michael Stone\n")
	chartFile.WriteString("% " + date + "\n")
	chartFile.WriteString("\n")
	chartFile.WriteString("# Notes [ ](data:tkt,mgr=&job=&want=)\n")
	chartFile.WriteString("\n")
	chartFile.WriteString("# Resume\n")
	chartFile.WriteString("\n")

	filepath.Walk(pdfPagesDir+"/",
		func(walkPath string, info os.FileInfo, walkErr error) error {
			return convertPdfPages(safeName, chart, chartFile, svgPagesDir, walkPath, info, walkErr)
		})

	return nil
}

var renameRegexp *regexp.Regexp = regexp.MustCompile("[^a-zA-Z-_\\/.]")

func SimplifyName(name string) string {
	return renameRegexp.ReplaceAllString(name, "")
}

func Convert(inputPath, outputPath, displayName string) error {
	inputName := path.Base(inputPath)
	safeName := SimplifyName(inputName)
	return convert(inputPath, safeName, outputPath, displayName)
}

var outputPath = flag.String("o", "./charts", "output dir")

func main() {
	flag.Parse()

	if outputPath == nil {
		glog.Fatalf("convert: job id warning: must set output path with -o")
	}

	chartsDir := *outputPath
	err := os.MkdirAll(chartsDir, 0755)
	if err != nil {
		L("convert: mkdir warning: %s", err)
		return
	}

	originalInputs := flag.Args()

	for _, inputPath := range originalInputs {
		displayName := SimplifyName(path.Base(inputPath))
		Convert(inputPath, *outputPath, displayName)
	}
}
