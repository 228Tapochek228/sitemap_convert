package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
)

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

type Node struct {
	Name     string
	Children map[string]*Node
	IsFile   bool
}

func main() {
	sitemapFile := flag.String("map", "sitemap.xml", "Path to sitemap.xml file")
	flag.Parse()

	xmlFile, err := os.Open(*sitemapFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var urlSet URLSet
	err = xml.Unmarshal(byteValue, &urlSet)
	if err != nil {
		fmt.Printf("Error parsing XML: %v\n", err)
		os.Exit(1)
	}

	root := &Node{
		Name:     "",
		Children: make(map[string]*Node),
		IsFile:   false,
	}

	// Построение дерева
	for _, u := range urlSet.URLs {
		parsedURL, err := url.Parse(u.Loc)
		if err != nil {
			continue
		}

		// Обрабатываем путь
		urlPath := parsedURL.Path
		if urlPath == "" || urlPath == "/" {
			root.IsFile = true // Корневой URL (сайт) считаем "файлом"
			continue
		}

		// Нормализуем путь
		urlPath = path.Clean(urlPath)
		if !strings.HasPrefix(urlPath, "/") {
			urlPath = "/" + urlPath
		}

		parts := strings.Split(urlPath, "/")[1:] // Игнорируем первый пустой элемент
		currentNode := root

		for i, part := range parts {
			if part == "" {
				continue
			}

			// Определяем, это файл или директория
			isFile := i == len(parts)-1 && strings.Contains(part, ".")

			// Создаем узел если его нет
			if _, exists := currentNode.Children[part]; !exists {
				currentNode.Children[part] = &Node{
					Name:     part,
					Children: make(map[string]*Node),
					IsFile:   isFile,
				}
			} else if isFile {
				// Если узел уже есть, но теперь это файл - обновляем статус
				currentNode.Children[part].IsFile = isFile
			}

			currentNode = currentNode.Children[part]
		}
	}

	// Вывод дерева
	printTree(root, "", true)
}

func printTree(node *Node, prefix string, isLast bool) {
	// Вывод текущего узла
	if node.Name != "" {
		fmt.Print(prefix)
		if isLast {
			fmt.Print("└── ")
		} else {
			fmt.Print("├── ")
		}

		fmt.Print(node.Name)
		if node.IsFile {
			fmt.Println()
		} else {
			fmt.Println("/")
		}
	}

	// Сортируем дочерние узлы: сначала директории, потом файлы
	children := make([]*Node, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, child)
	}

	sort.Slice(children, func(i, j int) bool {
		if children[i].IsFile == children[j].IsFile {
			return children[i].Name < children[j].Name
		}
		return !children[i].IsFile
	})

	// Рекурсивный вывод дочерних узлов
	for i, child := range children {
		newPrefix := prefix
		if node.Name != "" {
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
		}

		printTree(child, newPrefix, i == len(children)-1)
	}
}
