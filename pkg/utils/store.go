package utils

var (
	storeRegistry SyncMap[any, *SyncMap[string, any]]
)

func GetStore(name any) *SyncMap[string, any] {

	ok := storeRegistry.Has(name)
	if !ok {
		storeRegistry.Set(name, &SyncMap[string, any]{})
	}

	return storeRegistry.Get(name)
}

func DestroyStore(name any) {
	storeRegistry.Delete(name)
}

func Namespace(keys ...string) string {
	result := ""

	for idx, s := range keys {
		result += s
		if (len(keys) - 1) != idx {
			result += ":"
		}
	}

	return result
}
