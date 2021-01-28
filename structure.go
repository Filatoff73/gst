package gst

/*
#cgo pkg-config: gstreamer-1.0
#include "gst.h"
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type Structure struct {
	C *C.GstStructure
}

func NewStructure(name string) (structure *Structure) {
	CName := (*C.gchar)(unsafe.Pointer(C.CString(name)))
	CGstStructure := C.gst_structure_new_empty(CName)

	structure = &Structure{
		C: CGstStructure,
	}

	runtime.SetFinalizer(structure, func(structure *Structure) {
		C.gst_structure_free(structure.C)
	})

	return
}

func StructureFromString(structure string) (gstStruct *Structure) {
	c := (*C.gchar)(unsafe.Pointer(C.CString(structure)))
	defer C.g_free(C.gpointer(unsafe.Pointer(c)))
	CStruct := C.gst_structure_from_string(c,nil)
	gstStruct = &Structure{
		C: CStruct,
	}
	runtime.SetFinalizer(gstStruct, func(gstStruct *Structure) {
		C.gst_structure_free(gstStruct.C)
	})

	return
}

func (s *Structure) SetValue(name string, value interface{}) {

	CName := (*C.gchar)(unsafe.Pointer(C.CString(name)))
	defer C.g_free(C.gpointer(unsafe.Pointer(CName)))
	switch value.(type) {
	case string:
		str := (*C.gchar)(unsafe.Pointer(C.CString(value.(string))))
		defer C.g_free(C.gpointer(unsafe.Pointer(str)))
		C.X_gst_structure_set_string(s.C, CName, str)
	case int:
		C.X_gst_structure_set_int(s.C, CName, C.gint(value.(int)))
	case uint32:
		C.X_gst_structure_set_uint(s.C, CName, C.guint(value.(uint32)))
	case bool:
		var v int
		if value.(bool) == true {
			v = 1
		} else {
			v = 0
		}
		C.X_gst_structure_set_bool(s.C, CName, C.gboolean(v))
	}

	return
}

func (s *Structure) ToString() (str string) {
	Cstr := C.gst_structure_to_string(s.C)
	str = C.GoString((*C.char)(unsafe.Pointer(Cstr)))
	C.g_free((C.gpointer)(unsafe.Pointer(Cstr)))

	return
}

func (s *Structure) GetStructureName() (str string) {
	Cstr := C.gst_structure_get_name(s.C)
	str = C.GoString((*C.char)(unsafe.Pointer(Cstr)))
	return
}

func (s *Structure) GetDoubleArrayValue(key string) (res []float64,err error) {

	res = make([]float64,0)
	CKey := (*C.gchar)(unsafe.Pointer(C.CString(key)))
	defer C.g_free(C.gpointer(unsafe.Pointer(CKey)))

	Carray := C.gst_structure_get_value(s.C, CKey)
	if Carray==nil {
		err = errors.New("nil Carray")
		return
	}
	CKeyArray := (*C.GValueArray)(C.g_value_get_boxed(Carray))
	if CKeyArray==nil {
		err = errors.New("nil CKeyArray")
		return
	}

	size := CKeyArray.n_values
	if size==0 {
		err = errors.New("empty CKeyArray")
		return
	}
	var i C.uint
	for i=0;i<size;i++ {
		value := C.g_value_array_get_nth(CKeyArray,i);
		rms_dB := C.g_value_get_double (value);
		res = append(res,float64(rms_dB))
	}

	return
}
