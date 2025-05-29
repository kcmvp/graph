package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/tidwall/buntdb"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"sync"
)

const schemaPrefix = "_schema:"

var _ Graph[Vertex] = (*store[Vertex])(nil)
var db *buntdb.DB

func init() {
	var err error
	var u *user.User
	u, err = user.Current()
	if err != nil {
		panic(fmt.Errorf("can not get current user: %w", err))
	}
	dir := filepath.Join(u.HomeDir, "graph")
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		panic(fmt.Errorf("can not create directory: %w", err))
	}
	db, err = buntdb.Open(filepath.Join(dir, "db.data"))
	if err != nil {
		panic(fmt.Errorf("can not open database: %w", err))
	}
}

type store[T any] lo.Tuple3[*buntdb.DB, *Traits, sync.Map]

func newStore[T Vertex](options ...func(*Traits)) *store[T] {
	var p Traits
	for _, option := range options {
		option(&p)
	}
	if p.EdgeIndicator == 0 {
		p.EdgeIndicator = '-'
		log.Println("no edge separator settled，use default value '-'")
	}

	return &store[T]{A: db, B: &p, C: sync.Map{}}
}

func (s *store[T]) Traits() *Traits {
	return s.B
}

func (s *store[T]) AddVertex(value T) error {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) Vertex(hash string) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) Vertices() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) RemoveVertex(hash string) error {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) UpdateVertex(hash string, attributes map[rune]any) error {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) EdgeSeparator() rune {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) AddEdge(e Edge) error {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) Edge(hash string) (Edge, error) {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) Edges() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) UpdateEdge(hash string, attributes map[rune]any) error {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) RemoveEdge(hash string) error {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) Order() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *store[T]) Size() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (g *store[T]) register(v Vertex) (string, error) {
	concreteType := reflect.TypeOf(v)
	if concreteType.Kind() == reflect.Ptr {
		concreteType = concreteType.Elem() // 获取实际的结构体类型
	}

	// 1. 类型校验：确保是结构体类型
	if concreteType.Kind() != reflect.Struct {
		return "", fmt.Errorf("Vertex must be a struct, but it's %s", concreteType.Kind())
	}

	// 2. 尝试从内存 sync.Map 加载 (L1 缓存)
	if actual, ok := g.C.Load(concreteType); ok {
		return actual.(string), nil // 命中内存缓存，直接返回
	}

	// 3. 内存缓存未命中，尝试从 buntdb 加载 (L2 缓存/持久化层)
	dbKey := fmt.Sprintf("%s%s", schemaPrefix, concreteType.PkgPath()+"."+concreteType.Name())
	var loadedFullPath string

	err := g.A.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(dbKey)
		if err != nil {
			if errors.Is(err, buntdb.ErrNotFound) {
				return nil // buntdb 中也不存在
			}
			return err
		}
		loadedFullPath = val
		return nil
	})

	if err != nil && !errors.Is(err, buntdb.ErrNotFound) {
		return "", fmt.Errorf("从 buntdb 加载结构体全路径 '%s' 失败: %w", dbKey, err)
	}

	if loadedFullPath != "" {
		// 4. 从 buntdb 加载到内存缓存，并返回
		// 使用 LoadOrStore 确保线程安全：如果其他 goroutine 在此期间已经存储，则使用已存在的
		if actual, loaded := g.C.LoadOrStore(concreteType, loadedFullPath); loaded {
			return actual.(string), nil // 其他 goroutine 已存入，使用其值
		}
		return loadedFullPath, nil // 成功存入并返回
	}

	// 5. 内存和 buntdb 均未命中，计算并存储
	calculatedFullPath := concreteType.PkgPath() + "." + concreteType.Name()

	// 存储到 buntdb
	err = g.A.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(dbKey, calculatedFullPath, nil)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("将结构体全路径 '%s' 写入 buntdb 失败: %w", dbKey, err)
	}

	// 存储到内存 sync.Map
	// 再次使用 LoadOrStore 确保原子性和线程安全
	if actual, loaded := g.C.LoadOrStore(concreteType, calculatedFullPath); loaded {
		return actual.(string), nil // 其他 goroutine 已存入，使用其值
	}
	return calculatedFullPath, nil // 成功存入并返回
}

func marshalVertex(v Vertex) ([]byte, error) {
	abbreviatedMap := make(map[string]interface{})
	schema := v.Schema() // 获取节点的 Schema 映射

	// 通过反射获取 Vertex 实例的实际值
	concreteValue := reflect.ValueOf(v)
	if concreteValue.Kind() == reflect.Ptr {
		concreteValue = concreteValue.Elem() // 如果是指针，获取其指向的结构体值
	}
	if concreteValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("marshalVertex：Vertex 实现必须是结构体或结构体指针，但得到 %s", concreteValue.Kind())
	}

	// 遍历 Schema 映射，将结构体字段的值填充到缩写后的 map 中
	for fullPropertyName, shortRune := range schema {
		fieldValue := concreteValue.FieldByName(fullPropertyName) // 获取指定字段的值

		// 检查字段是否有效（存在且可导出）
		if !fieldValue.IsValid() || !fieldValue.CanInterface() {
			fmt.Printf("警告：在 %T 中未找到或无法导出字段 '%s'，已跳过。\n", v, fullPropertyName)
			continue
		}
		abbreviatedMap[string(shortRune)] = fieldValue.Interface() // 将字段值存入缩写 map
	}

	// 将缩写 map 序列化为 JSON 字节
	return json.Marshal(abbreviatedMap)
}

func unmarshalVertex(data []byte, target Vertex) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("unmarshalVertex：target 必须是一个非空的 Vertex 实现指针")
	}

	concreteValue := targetValue.Elem() // 获取指针指向的实际结构体值
	if concreteValue.Kind() != reflect.Struct {
		return fmt.Errorf("unmarshalVertex：target 必须是一个结构体指针，但得到 %s", concreteValue.Kind())
	}

	schema := target.Schema() // 获取目标实例的 Schema 映射
	shortToFullMap := make(map[rune]string)
	// 反转 Schema 映射，以便通过缩写键查找完整字段名
	for full, short := range schema {
		shortToFullMap[short] = full
	}

	tempMap := make(map[string]interface{})
	// 将 JSON 数据反序列化到临时 map 中
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	// 遍历临时 map，将值设置回目标结构体
	for shortRuneStr, value := range tempMap {
		shortRune := rune(shortRuneStr[0]) // 获取缩写字符

		fullPropertyName, ok := shortToFullMap[shortRune]
		if !ok {
			fmt.Printf("警告：在 %T 的 JSON 中发现未注册的缩写名 '%c'，已跳过。\n", target, shortRune)
			continue
		}

		fieldToSet := concreteValue.FieldByName(fullPropertyName) // 获取要设置的字段
		if !fieldToSet.IsValid() || !fieldToSet.CanSet() {
			fmt.Printf("警告：在 %T 中未找到或无法设置字段 '%s'，已跳过。\n", target, fullPropertyName)
			continue
		}

		// 将值重新序列化再反序列化，以处理 JSON unmarshal 的类型推断问题
		valueBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("为字段 %s 序列化值失败: %w", fullPropertyName, err)
		}
		if err := json.Unmarshal(valueBytes, fieldToSet.Addr().Interface()); err != nil {
			return fmt.Errorf("将值反序列化到字段 %s (%s) 失败: %w", fullPropertyName, fieldToSet.Type(), err)
		}
	}

	return nil
}
