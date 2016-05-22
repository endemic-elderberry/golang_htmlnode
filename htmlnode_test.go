package HTMLNode


import "testing"
import "github.com/stretchr/testify/assert"
import "golang.org/x/net/html"
import "strings"
import "regexp"

const SpacedHTMLData = `
<!doctype html>
<html>
  <head>
  </head>
  <body>
    Yo
    <div id="whatlist">
      <div class="what"></div>
      <div class="what"></div>
      <div class="what"></div>
      <div class="what"></div>
    </div>
  </body>
</html>`;

var HTMLData string;


func generateTestTree() HTMLNode {
  HTMLData = SpacedHTMLData;
  spaceReg := regexp.MustCompile(`[ \t]*\n[ \t]*`)
  dontReplace := false
  if !dontReplace {
    HTMLData = spaceReg.ReplaceAllString(SpacedHTMLData, "")
  }
  
  doc, err := html.Parse(writeHTML{HTMLData})
  if err != nil {
    panic(err)
  }
  return HTMLNodeForNode(doc)
}

func TestSetAttribute(t *testing.T) {
  assert := assert.New(t)
  
  hn := CreateElement("div")
  hn.SetAttribute("data-reactid", "something")
  hn.SetAttribute("data-hello", "hi")
  
  asHtml := hn.OuterHTML()
  
  expectedHTML := `<div data-reactid="something" data-hello="hi"></div>`
  
  assert.Equal(expectedHTML, asHtml, "SetAttribute works")
  assert.Equal(hn.GetAttribute("data-reactid"), "something")
  assert.Equal(hn.GetAttribute("data-hello"), "hi")
  
  assert.Panics(func () {
    hn.SetAttribute("data>yo", "value")
  },"Illegal attribute name")
}
func TestTraverseNodes(t *testing.T) {
  assert := assert.New(t)
  
  hn := generateTestTree()
  
  n := 0
  nodeTags := make([]string, 0)
  nodeTypes := make([]int, 0)
  nodeText := make([]string, 0)
  traverseNodes(hn, func (subHN HTMLNode) {
    nodeTypes = append(nodeTypes, subHN.NodeType())
    nodeTags = append(nodeTags, subHN.TagName())
    nodeText = append(nodeText, subHN.TextContent())
    n += 1
  });
  
  assert.Equal(10, n, "Should have 10 nodes")
}
func TestQuerySelector(t *testing.T) {
  
  assert := assert.New(t)
  
  hn := generateTestTree()
  head := hn.QuerySelector("head");
  
  assert.Equal(head.TagName(), "head", "is a head")
  
  headOtherWay := hn.ChildNode(1).ChildNode(0)
  assert.Equal(head, headOtherWay, "same head")
}
func TestQuerySelectorAll(t *testing.T) {
  
  assert := assert.New(t)
  
  hn := generateTestTree()
  whatDivs := hn.QuerySelectorAll(".what");
  
  assert.Equal(len(whatDivs), 4, "Found 4 divs")
}
func TestHTMLNodeOuterHTML(t *testing.T) {
  assert := assert.New(t)
  
  hn := generateTestTree()
  outerHTML := hn.OuterHTML()
  
  assert.Equal(HTMLData, outerHTML, "Same as original HTMLData")
}

func TestBadQuery(t *testing.T) {
  assert := assert.New(t)
  
  hn := generateTestTree()
  
  whatList := hn.QuerySelector(".whatlist")
  
  assert.Nil(whatList, "is null")
}
func TestNewChild(t *testing.T) {
  assert := assert.New(t)
  
  hn := generateTestTree()
  
  whatList := hn.QuerySelector("#whatlist")
  
  existingLength := len(whatList.ChildNodes())
  
  what := CreateElement("div")
  what.SetAttribute("class", "what else")
  whatList.AppendChild(what)
  
  newLength :=  len(whatList.ChildNodes())
  
  foundNewItem := hn.QuerySelector(".what.else")
  assert.NotNil(foundNewItem, "what else not found")
  assert.Equal(newLength, existingLength + 1, "Has one more child")
  
}


func TestSetInnerHTMLEmpty(t *testing.T) {
  assert := assert.New(t)
  
  hn := generateTestTree()
  
  list := hn.QuerySelector("#whatlist")
  list.SetInnerHTML("hello")
  
  whatChildren := list.QuerySelectorAll(".what")
  
  assert.Equal(0, len(whatChildren), "Nothing should be left")
}

func TestSetInnerHTMLTag(t *testing.T) {
  assert := assert.New(t)
  
  hn := generateTestTree()
  
  list := hn.QuerySelector("#whatlist")
  list.SetInnerHTML("<span class=\"hi\"></span><span class=\"hi\"></span><span class=\"hi\"></span>")
  
  hiChildren := list.QuerySelectorAll(".hi")
  
  assert.Equal(3, len(hiChildren), "Should have 3 hi span now")
}


func TestHTMLNodeFromNode(t *testing.T) {
  assert := assert.New(t)
  
  hn := generateTestTree()
  
  childNodes := hn.ChildNodes()
  
  assert.Equal(2, hn.NodeType(), "Document node top")
  assert.Equal(2, len(childNodes), "have two children doctype and html")
  
  assert.NotNil(hn, "should successfully clonenode")
  
  textContent := hn.TextContent()
  
  textContent = strings.Trim(textContent, " \t\n")
  
  assert.Equal(textContent, "Yo", "Correct Trimmed TextContent")
}

