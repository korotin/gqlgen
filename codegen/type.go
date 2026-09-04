package codegen

import (
	"fmt"

	"github.com/99designs/gqlgen/codegen/config"
)

func (b *builder) buildTypes() map[string]*config.TypeReference {
	ret := map[string]*config.TypeReference{}
	for _, ref := range b.Binder.References {
		processType(ret, ref)
	}
	return ret
}

func processType(ret map[string]*config.TypeReference, ref *config.TypeReference) {
	key := ref.UniquenessKey()
	if existing, found := ret[key]; found {
		// Simplistic check of content which is obviously different.
		existingGQL := fmt.Sprintf("%v", existing.GQL)
		newGQL := fmt.Sprintf("%v", ref.GQL)
		if existingGQL != newGQL {
			panic(
				fmt.Sprintf(
					"non-unique key \"%s\", trying to replace %s with %s",
					key,
					existingGQL,
					newGQL,
				),
			)
		}
		// References sharing a key produce a single pair of marshal functions, so
		// the stored one has to cover every position any of them is reached from.
		ref.Directions |= existing.Directions
	}
	ret[key] = ref

	if ref.IsSlice() || ref.IsPtrToSlice() || ref.IsPtrToPtr() || ref.IsPtrToIntf() {
		processType(ret, ref.Elem())
	}
}

// RequireUnmarshal marks ref as reached from an input position, both on ref
// itself and on the reference the generator emits functions for. A plugin that
// unmarshals a reference it borrowed from an output position (federation reads
// @key and @requires fields back out of an entity representation) must call it
// from GenerateCode, or the unmarshaler it calls will not be generated.
func (d *Data) RequireUnmarshal(ref *config.TypeReference) {
	for ; ref != nil; ref = ref.Elem() {
		ref.Directions |= config.RefInput
		if stored := d.ReferencedTypes[ref.UniquenessKey()]; stored != nil {
			stored.Directions |= config.RefInput
		}
		if !ref.IsSlice() && !ref.IsPtrToSlice() && !ref.IsPtrToPtr() && !ref.IsPtrToIntf() {
			return
		}
	}
}
