package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	
	//"net/url"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
	//"github.com/mafredri/cdp/protocol/target"
	//"github.com/mafredri/cdp/protocol/network"
)

func main() {
	err := run(5 * time.Second)
	//a,err:=CreatePdf(5 * time.Second,"http://google.com", 10.0, 10.0)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(a)
}

func run(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New("http://127.0.0.1:9222")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return err
		}
	}

	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return err
	}
	defer conn.Close() // Leaving connections open will leak memory.

	c := cdp.NewClient(conn)

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = c.Page.Enable(ctx); err != nil {
		return err
	}
	
	// go is musch faer than usual, soo sleeping it for sometime ;
    time.Sleep(10)
    
	// Create the Navigate arguments with the optional Referrer field set.
	navArgs := page.NewNavigateArgs("https://arshpreetsingh.github.io/").
		SetReferrer("https://duckduckgo.com")
	nav, err := c.Page.Navigate(ctx, navArgs)
	if err != nil {
		return err
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		return err
	}

	fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	doc, err := c.DOM.GetDocument(ctx, nil)
	if err != nil {
		return err
	}

	// Get the outer HTML for the page.
	result, err := c.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &doc.Root.NodeID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("HTML: %s\n", result.OuterHTML)

	// Capture a screenshot of the current page.
	screenshotName := "screenshot.jpg"
	screenshotArgs := page.NewCaptureScreenshotArgs().
		SetFormat("jpeg").
		SetQuality(80)
	screenshot, err := c.Page.CaptureScreenshot(ctx, screenshotArgs)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(screenshotName, screenshot.Data, 0644); err != nil {
		return err
	}

	fmt.Printf("Saved screenshot: %s\n", screenshotName)
	
	// do a PDF print as well 
	
	printToPDFArgs := page.NewPrintToPDFArgs().
		SetLandscape(true).
		SetPrintBackground(true).
		SetMarginTop(0).
		SetMarginBottom(0).
		SetMarginLeft(0).
		SetMarginRight(0).
		SetPrintBackground(true).
		SetPaperWidth(10).
		SetPaperHeight(10)
		
	pdfFile, _ := c.Page.PrintToPDF(ctx, printToPDFArgs)
	
    //fmt.Println(print.Data)
    if err = ioutil.WriteFile("file.pdf", pdfFile.Data, 0644); err != nil {
		return err
	}
	
    
//    permissions := 0644 // or whatever you need
//byteArray := []byte("to be written to a file\n")
//err := ioutil.WriteFile("file.txt", byteArray, permissions)
//if err != nil { 
    // handle error
//}

	return nil
}
