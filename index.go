package golang_htmlnode

import "github.com/andybalholm/cascadia"
import "golang.org/x/net/html"
import "fmt"
import "io"

type HTMLNode interface{
  GetAttribute(string) string
  SetAttribute(string,string) bool
  TextContent() string
  ChildNodes() []HTMLNode
  SetParentNode(HTMLNode)bool
  FirstChild() HTMLNode
  ChildNode(int)HTMLNode
  LastChild() HTMLNode
  NodeType() int
  InnerHTML() string
  SetInnerHTML(string)
  OuterHTML() string
  TagName() string
  ParentNode() HTMLNode
  MirrorNode() *html.Node
  QuerySelector(string)HTMLNode
  QuerySelectorAll(string)[]HTMLNode
  AppendChild(HTMLNode)bool
  RemoveChild(HTMLNode)bool
}


type writeHTML struct{
  ToWrite string
}
type hNodeStruct struct{
  FieldParentNode HTMLNode
  FieldTagName string
  FieldNodeType int
  TextData string
  FieldChildNodes []HTMLNode
  AttrOrder []string
  AttrIndex map[string]int
  AttrMap map[string]string
  FieldMirrorNode *html.Node
}
func (node *hNodeStruct) SetParentNode(pNode HTMLNode) bool {
  mirror := node.FieldMirrorNode
  
  if pNode == nil {
    node.FieldParentNode = nil
    mirror.Parent = nil
    mirror.PrevSibling = nil
    mirror.NextSibling = nil
    return true
  }
  if node.FieldParentNode != nil {
    return false
  }
  
  pMirror:= pNode.MirrorNode()
  
  if mirror.Parent != pMirror {
    mirror.Parent = pMirror
  }
  
  node.FieldParentNode = pNode
  return true;
}
func (node *hNodeStruct) ParentNode() HTMLNode {
  return node.FieldParentNode
}
func (node *hNodeStruct) MirrorNode() *html.Node {
  return node.FieldMirrorNode
}
func (node *hNodeStruct) RemoveChild(cNode HTMLNode) bool {
  nodeFound := false;
  var matchedNode HTMLNode;
  var index int;
  for i := 0 ; i < len(node.FieldChildNodes) ; i++ {
    subCNode := node.FieldChildNodes[i]
    if cNode == subCNode {
      index = 0;
      nodeFound = true;
      matchedNode = subCNode;
    }
  }
  if !nodeFound {
    return false;
  }
  
  result := matchedNode.SetParentNode(nil)
  if !result {
    return false
  }
  before := node.FieldChildNodes[:index]
  after := node.FieldChildNodes[index + 1:]
  
  var fAfterMirror *html.Node;
  var lBeforeMirror *html.Node;
  fAfterMirror = nil
  lBeforeMirror= nil
  if len(after) > 0 {
    fAfterMirror = after[0].MirrorNode()
  }
  if len(before) > 0 {
    lBeforeMirror = before[len(before) - 1].MirrorNode()
  }
  
  node.FieldChildNodes = append(before, after...)
  
  mirror := node.FieldMirrorNode
  lastIndex := len(node.FieldChildNodes) - 1
  
  mirror.FirstChild = node.FieldChildNodes[0].MirrorNode()
  mirror.LastChild = node.FieldChildNodes[lastIndex].MirrorNode()
  
  if fAfterMirror != nil {
    fAfterMirror.NextSibling = lBeforeMirror
  }
  if lBeforeMirror!= nil {
    lBeforeMirror.PrevSibling = fAfterMirror
  }
  
  return true;
}
func (node *hNodeStruct) GetAttribute(key string) string {
  val, _ := node.AttrMap[key]
  return val
}
func (node *hNodeStruct) SetAttribute(key string, val string) bool {
  previousIndex, existed := node.AttrIndex[key]
  mirror := node.FieldMirrorNode
  attrList := mirror.Attr
  if !existed {
    
    escapedKey := html.EscapeString(key)
    if escapedKey != key {
      panic("illegal html attribute key")
    }
    
    attribute := html.Attribute{
      Key: key,
      Val: val,
    }
    newIndex := len(attrList)
    attrList = append(attrList, attribute)
    mirror.Attr = attrList
    
    node.AttrOrder = append(node.AttrOrder, key)
    node.AttrIndex[ key ] = newIndex
  } else {
    existingAttr := attrList[previousIndex]
    existingAttr.Val = val
    attrList[previousIndex] = existingAttr
    mirror.Attr = attrList
  }
  
  node.AttrMap[key] = val
  
  return true
}
func (node *hNodeStruct) RemoveAttribute(key string) bool {
  previousIndex, existed := node.AttrIndex[key]
  if !existed {
    return true
  }
  
  mirror := node.FieldMirrorNode
  attrList := mirror.Attr
  beforeAlist := attrList[:previousIndex]
  afterAlist := attrList[previousIndex + 1:]
  
  beforeOrder := node.AttrOrder[:previousIndex]
  afterOrder := node.AttrOrder[previousIndex + 1:]
  
  for i := 0 ; i < len(afterOrder) ; i++ {
    afterKey := afterOrder[i]
    prevLocation, prevExists := node.AttrIndex[afterKey]
    if !prevExists {
      panic("attribute already removed?")
    }
    newLocation := prevLocation - 1
    if newLocation < 0 {
      panic("illegal new newLocation?")
    }
    node.AttrIndex[afterKey] = newLocation
  }
  
  delete(node.AttrIndex, key)
  delete(node.AttrMap, key)
  node.AttrOrder = append(beforeOrder, afterOrder...)
  mirror.Attr = append(beforeAlist, afterAlist...)
  return true
}
func (node *hNodeStruct) NodeType() int {
  return node.FieldNodeType;
}
func (node *hNodeStruct) TagName() string {
  return node.FieldTagName;
}
func (node *hNodeStruct) FirstChild() HTMLNode {
  return node.FieldChildNodes[0]
}
func (node *hNodeStruct) ChildNode(index int) HTMLNode {
  return node.FieldChildNodes[ index ]
}
func (node *hNodeStruct) InnerHTML() string {
  s := ""
  for i := 0 ; i < len(node.FieldChildNodes) ; i++ {
    cNode := node.FieldChildNodes[i]
    s += cNode.OuterHTML()
  }
  return s
}
func (node *hNodeStruct) SetInnerHTML(htmlInput string) {
  cNodesCopy := node.FieldChildNodes[:]
  for i := 0 ; i < len(cNodesCopy) ; i++ {
    result := cNodesCopy[i].SetParentNode(nil)
    if !result {
      panic("not my child?")
    }
  }
  node.FieldChildNodes = make([]HTMLNode, 0)
  mNode := node.MirrorNode()
  mNode.FirstChild = nil
  mNode.LastChild = nil
  
  newChildren, err := html.ParseFragment(writeHTML{htmlInput}, node.MirrorNode())
  
  if err != nil {
    panic(err)
  }
  for i := 0 ; i < len(newChildren) ; i++ {
    node.AppendChild(HTMLNodeForNode(newChildren[i]))
  }
}
func (writer writeHTML) Read(bData []byte) (int, error) {
  bAppend := []byte(writer.ToWrite)
  
  for i := 0 ; i < len(bAppend) ; i++ {
    bData[i] = bAppend[i]
  }
  return len(writer.ToWrite),io.EOF
}
func findRelation(ancestor *html.Node, descendant  *html.Node) []int {
  relation := make([]int, 0)
  currentNode := descendant
  for {
    parent := currentNode.Parent
    if parent == nil {
      panic("not found")
    }
    iCount := 0
    for child:=parent.FirstChild; child != nil ; child = child.NextSibling {
      if child == currentNode {
        break;
      }
      iCount ++
    }
    relation = append([]int{iCount}, relation...)
    currentNode = parent
    if parent == ancestor {
      break;
    }
  }
  return relation
}
func (node *hNodeStruct) QuerySelectorAll(query string) []HTMLNode {
  
  queryObject := cascadia.MustCompile(query)
  mirror := node.FieldMirrorNode
  
  matchedList := queryObject.MatchAll(node.MirrorNode())
  
  nodeList := make([]HTMLNode, 0)
  for i := 0 ; i < len(matchedList) ; i++ {
    if mirror == matchedList[i] {
      nodeList = append(nodeList, node)
      continue
    }
    relation := findRelation(mirror, matchedList[i])
    var outputNode HTMLNode;
    outputNode = node;
    for i := 0 ; i < len(relation) ; i++ {
      outputNode = outputNode.ChildNode(relation[i])
    }
    nodeList = append(nodeList, outputNode)
  }
  return nodeList
}
func (node *hNodeStruct) QuerySelector(query string) HTMLNode {
  
  queryObject := cascadia.MustCompile(query)
  mirror := node.FieldMirrorNode
  
  matched := queryObject.MatchFirst(node.MirrorNode())
  if matched == nil {
    return nil
  }
  if matched == mirror {
    return node
  }
  relation := findRelation(mirror, matched)
  var outputNode HTMLNode;
  outputNode = node;
  for i := 0 ; i < len(relation) ; i++ {
    outputNode = outputNode.ChildNode(relation[i])
  }
  return outputNode
}
func (node *hNodeStruct) OuterHTML() string {
  switch node.FieldNodeType {
    case 1:
      return html.EscapeString(node.TextData)
    case 5:
      return "<!doctype html>"
    case 2, 3:
      s := ""
      if node.FieldNodeType == 3 {
        s += "<" + node.TagName()
        for _, k := range node.AttrOrder {
          v, vok := node.AttrMap[k]
          if !vok {
            panic("missing attr?")
          }
          s += fmt.Sprintf(" %s=\"%s\"",k, html.EscapeString(v))
        }
        s += ">"
      }
      s += node.InnerHTML()
      if node.FieldNodeType == 3 {
        s += "</" + node.TagName() + ">"
      }
      return s
    default:
      panic(fmt.Sprintf("unhandled %d nodeType", node.FieldNodeType))
  }
}
func (node *hNodeStruct) LastChild() HTMLNode {
  lindex := len(node.FieldChildNodes) - 1
  if lindex >= 0 {
    return node.FieldChildNodes[lindex]
  }
  return nil
}
func (node *hNodeStruct) AppendChild(hn HTMLNode) bool {
  switch node.NodeType() {
    case 2, 3:
      break;
    default:
      panic("This node type does not support this method")
  }
  
  result := hn.SetParentNode(node)
  if ! result {
    return false
  }
  node.FieldChildNodes = append(node.FieldChildNodes, hn)
  
  mirror := node.FieldMirrorNode
  cMirror:= hn.MirrorNode()
  
  lChild := mirror.LastChild
  if lChild == nil {
    mirror.FirstChild = cMirror
    mirror.LastChild = cMirror
    cMirror.NextSibling = nil
    cMirror.PrevSibling = nil
    cMirror.Parent = mirror
    return true;
  }
  
  lChild.NextSibling = cMirror
  cMirror.PrevSibling = lChild
  mirror.LastChild = cMirror
  
  return true
}
func (node *hNodeStruct) ChildNodes() []HTMLNode {
  return node.FieldChildNodes[:]
}
func traverseNodes(node HTMLNode, each_node func(HTMLNode)) {
  childNodes := node.ChildNodes()
  for i := 0 ; i < len(childNodes) ; i++ {
    each_node(childNodes[i])
    traverseNodes(childNodes[i], each_node)
  }
}
func (node *hNodeStruct) TextContent() string {
  nodeType := node.NodeType()
  switch nodeType {
    case 1:
      return node.TextData
    case 5:
      return ""
    case 3, 2:
      s := ""
      childNodes := node.ChildNodes()
      for i := 0 ; i < len(childNodes) ; i++ {
        s += childNodes[i].TextContent()
      }
      return s
    default:
      panic(fmt.Sprintf("undefined nodeType %d", nodeType))
  }
  return ""
}
func CreateTextNode(txt string) HTMLNode {
  mirror := &html.Node{
    Data: txt,
    Type: 1,
  }
  return &hNodeStruct{
    FieldNodeType: 1,
    TextData: txt,
    FieldMirrorNode: mirror,
  }
}
func CreateElement(tag string) HTMLNode {
  mirror := &html.Node{
    Data: tag,
    Type: 3,
    Attr: make([]html.Attribute, 0),
  }
  return &hNodeStruct{
    AttrOrder: make([]string, 0),
    AttrMap: make(map[string]string),
    AttrIndex: make(map[string]int),
    FieldNodeType: 3,
    FieldTagName : tag,
    FieldMirrorNode: mirror,
  };
}

func HTMLNodeForNode(node *html.Node) HTMLNode {
  nType := int(node.Type)
  
  newNode := &hNodeStruct{
    FieldNodeType: nType,
    FieldMirrorNode: node,
  }
  
  if nType != 1 {
    AttrMap := make(map[string]string)
    AttrIndex:=make(map[string]int)
    AttrOrder:= make([]string, 0)
    
    for i := 0 ; i < len(node.Attr) ; i++ {
      attr := node.Attr[i]
      AttrMap[attr.Key] = attr.Val
      AttrIndex[attr.Key] = i
      AttrOrder = append(AttrOrder, attr.Key)
    }
    newNode.AttrOrder = AttrOrder
    newNode.AttrIndex = AttrIndex
    newNode.AttrMap = AttrMap
  }
  
  var textData string;
  var childNodes []HTMLNode;
  switch nType {
    case 3, 2:
      switch nType {
        case 3:
          tagName := node.Data
          escapedTagName := html.EscapeString(tagName)
          if escapedTagName != tagName {
            panic("illegal tag name " + tagName)
          }
          newNode.FieldTagName = tagName
      }
      childNodes = make([]HTMLNode, 0)
      for child:=node.FirstChild ; child != nil ; child = child.NextSibling {
        newChild := HTMLNodeForNode(child)
        result := newChild.SetParentNode(newNode)
        if !result {
          panic("Failed to set parent node")
        }
        childNodes = append(childNodes, newChild)
      }
    case 1:
      textData = node.Data;
    case 5:
      break
    default:
      panic(fmt.Sprintf("unhandled nodeType %d", nType))   
  }
  newNode.FieldChildNodes = childNodes
  newNode.TextData = textData
  
  return newNode
}
func ParseReader(writer io.Reader) (HTMLNode, error) {
  doc, err := html.Parse(writer)
  if err != nil {
    return nil,err
  }
  return HTMLNodeForNode(doc),err
}
func Parse(html string) (HTMLNode, error) {
  return ParseReader(writeHTML{html})
}

