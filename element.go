package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
#include "gst.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

var (
	mutex         sync.Mutex
	callbackStore = map[uint64]*Element{}
)

type PadAddedCallback func(element *Element, pad *Pad)

type State C.GstState

const (
	StateVoidPending = State(C.GST_STATE_VOID_PENDING)
	StateNull         = State(C.GST_STATE_NULL)
	StateReady        = State(C.GST_STATE_READY)
	StatePaused       = State(C.GST_STATE_PAUSED)
	StatePlaying      = State(C.GST_STATE_PLAYING)
)

func (s *State) g() *C.GstState {
	return (*C.GstState)(s)
}

func (s State) String() string {
	switch s {
	case StateVoidPending:
		return "STATE_VOID_PENDING"
	case StateNull:
		return "STATE_NULL"
	case StateReady:
		return "STATE_READY"
	case StatePaused:
		return "STATE_PAUSED"
	case StatePlaying:
		return "STATE_PLAYING"
	}
	return "Unknown state"
}

type StateChangeReturn C.GstStateChangeReturn

const (
	STATE_CHANGE_FAILURE    = StateChangeReturn(C.GST_STATE_CHANGE_FAILURE)
	STATE_CHANGE_SUCCESS    = StateChangeReturn(C.GST_STATE_CHANGE_SUCCESS)
	STATE_CHANGE_ASYNC      = StateChangeReturn(C.GST_STATE_CHANGE_ASYNC)
	STATE_CHANGE_NO_PREROLL = StateChangeReturn(C.GST_STATE_CHANGE_NO_PREROLL)
)

type Element struct {
	GstElement *C.GstElement
	onPadAdded PadAddedCallback
	callbackID uint64
}

func (e *Element) Name() (name string) {
	n := (*C.char)(unsafe.Pointer(C.gst_object_get_name((*C.GstObject)(unsafe.Pointer(e.GstElement)))))
	if n != nil {
		name = string(nonCopyCString(n, C.int(C.strlen(n))))
	}

	return
}

func (e *Element) Link(dst *Element) bool {

	result := C.gst_element_link(e.GstElement, dst.GstElement)
	if result == C.TRUE {
		return true
	}
	return false
}


func (e *Element) LinkFileterd(dst *Element,gstCaps *Caps) bool {
	result := C.gst_element_link_filtered(e.GstElement, dst.GstElement,gstCaps.caps)
	if result == C.TRUE {
		return true
	}
	return false
}

func (e *Element) UnLink(dst *Element)  {
	C.gst_element_unlink(e.GstElement, dst.GstElement)
}

func (e *Element) GetPadTemplate(name string) (padTemplate *PadTemplate) {

	n := (*C.gchar)(unsafe.Pointer(C.CString(name)))
	defer C.g_free(C.gpointer(unsafe.Pointer(n)))
	CPadTemplate := C.gst_element_class_get_pad_template(C.X_GST_ELEMENT_GET_CLASS(e.GstElement), n)
	padTemplate = &PadTemplate{
		C: CPadTemplate,
	}

	return
}

func (e *Element) GetRequestPad(padTemplate *PadTemplate, name string, caps *Caps) (pad *Pad) {

	var n *C.gchar
	var c *C.GstCaps

	if name == "" {
		n = nil
	} else {
		n = (*C.gchar)(unsafe.Pointer(C.CString(name)))
		defer C.g_free(C.gpointer(unsafe.Pointer(n)))
	}
	if caps == nil {
		c = nil
	} else {
		c = caps.caps
	}
	CPad := C.gst_element_request_pad(e.GstElement, padTemplate.C, n, c)
	pad = &Pad{
		pad: CPad,
	}

	return
}

//todo fail build on buildserver vendor/github.com/Filatoff73/gst/element.go:158:52: constant 18446744073709551615 overflows _Ctype_long
//func (e *Element) SeekSimple(millisec int64) bool  {
//	res:= C.gst_element_seek_simple(e.GstElement, C.GST_FORMAT_TIME,
//		C.GST_SEEK_FLAG_FLUSH, C.gint64(millisec*1000000))
//	if res>0 {
//		return true
//	}
//	return false
//}
//
//func (e *Element) Seek(millisec int64) bool  {
//	//res:= C.gst_element_seek(e.GstElement,
//	//	1.0,
//	//	C.GST_FORMAT_TIME,
//	//	C.GST_SEEK_FLAG_FLUSH,
//	//	C.GST_SEEK_TYPE_SET, C.gint64(millisec*1000000), C.GST_SEEK_TYPE_NONE, C.GST_CLOCK_TIME_NONE)
//
//	res:= C.gst_element_seek(e.GstElement,
//		1.0,
//		C.GST_FORMAT_TIME,
//		 C.GST_SEEK_FLAG_FLUSH,
//		C.GST_SEEK_TYPE_SET, C.gint64(millisec*1000000), C.GST_SEEK_TYPE_NONE, C.GST_CLOCK_TIME_NONE)
//	if res>0 {
//		return true
//	}
//
//	return false
//}




func (e *Element) GetStaticPad(name string) (pad *Pad) {

	n := (*C.gchar)(unsafe.Pointer(C.CString(name)))
	defer C.g_free(C.gpointer(unsafe.Pointer(n)))
	CPad := C.gst_element_get_static_pad(e.GstElement, n)
	pad = &Pad{
		pad: CPad,
	}

	return
}

func (e *Element) GetRequestPad2(name string) (pad *Pad) {

	n := (*C.gchar)(unsafe.Pointer(C.CString(name)))
	defer C.g_free(C.gpointer(unsafe.Pointer(n)))
	CPad := C.gst_element_get_request_pad(e.GstElement, n)
	pad = &Pad{
		pad: CPad,
	}

	return
}

func (e *Element) AddPad(pad *Pad) bool {

	Cret := C.gst_element_add_pad(e.GstElement, pad.pad)
	if Cret == 1 {
		return true
	}

	return false
}

func (e *Element) RemovePad(pad *Pad) bool {

	Cret := C.gst_element_remove_pad(e.GstElement, pad.pad)
	if Cret == 1 {
		return true
	}

	return false
}


func (e *Element) SetState(state State) StateChangeReturn {
	return StateChangeReturn(C.gst_element_set_state(e.GstElement, C.GstState(state)))
}

func (e *Element) GetState(timeout_ns int64) (state, pending State,
	ret StateChangeReturn) {
	ret = StateChangeReturn(C.gst_element_get_state(
		e.GstElement, state.g(), pending.g(), C.GstClockTime(timeout_ns),
	))
	return
}

func (e *Element) GetClockBaseTime() uint64 {

	CClockTime := C.gst_element_get_base_time(e.GstElement)

	return uint64(CClockTime)
}

func (e *Element) GetClock() (gstClock *Clock) {

	CElementClock := C.gst_element_get_clock(e.GstElement)

	gstClock = &Clock{
		C: CElementClock,
	}

	runtime.SetFinalizer(gstClock, func(gstClock *Clock) {
		C.gst_object_unref(C.gpointer(unsafe.Pointer(gstClock.C)))
	})

	return
}

func (e *Element) PushBuffer(data []byte) (err error) {

	b := C.CBytes(data)
	defer C.free(b)

	var gstReturn C.GstFlowReturn

	gstReturn = C.X_gst_app_src_push_buffer(e.GstElement, b, C.int(len(data)))

	if gstReturn != C.GST_FLOW_OK {
		err = errors.New("could not push buffer on appsrc element")
		return
	}

	return
}

func (e *Element) PullSample() (sample *Sample, err error) {

	CGstSample := C.gst_app_sink_pull_sample((*C.GstAppSink)(unsafe.Pointer(e.GstElement)))
	if CGstSample == nil {
		err = errors.New("could not pull a sample from appsink")
		return
	}

	gstBuffer := C.gst_sample_get_buffer(CGstSample)

	if gstBuffer == nil {
		err = errors.New("could not pull a sample from appsink")
		return
	}

	mapInfo := (*C.GstMapInfo)(unsafe.Pointer(C.malloc(C.sizeof_GstMapInfo)))
	defer C.free(unsafe.Pointer(mapInfo))

	if int(C.X_gst_buffer_map(gstBuffer, mapInfo)) == 0 {
		err = errors.New(fmt.Sprintf("could not map gstBuffer %#v", gstBuffer))
		return
	}

	CData := (*[1 << 30]byte)(unsafe.Pointer(mapInfo.data))
	data := make([]byte, int(mapInfo.size))
	copy(data, CData[:])

	duration := uint64(C.X_gst_buffer_get_duration(gstBuffer))

	sample = &Sample{
		Data:     data,
		Duration: duration,
	}

	C.gst_buffer_unmap(gstBuffer, mapInfo)
	C.gst_sample_unref(CGstSample)

	return
}

// appsink
func (e *Element) IsEOS() bool {

	Cbool := C.gst_app_sink_is_eos((*C.GstAppSink)(unsafe.Pointer(e.GstElement)))
	if Cbool == 1 {
		return true
	}

	return false
}

func (e *Element) SetObject(name string, value interface{}) {

	cname := (*C.gchar)(unsafe.Pointer(C.CString(name)))
	defer C.g_free(C.gpointer(unsafe.Pointer(cname)))
	switch value.(type) {
	case string:
		str := (*C.gchar)(unsafe.Pointer(C.CString(value.(string))))
		defer C.g_free(C.gpointer(unsafe.Pointer(str)))
		C.X_gst_g_object_set_string(e.GstElement, cname, str)
	case int:
		C.X_gst_g_object_set_int(e.GstElement, cname, C.gint(value.(int)))
	case uint32:
		C.X_gst_g_object_set_uint(e.GstElement, cname, C.guint(value.(uint32)))
	case bool:
		var cvalue int
		if value.(bool) == true {
			cvalue = 1
		} else {
			cvalue = 0
		}
		C.X_gst_g_object_set_bool(e.GstElement, cname, C.gboolean(cvalue))
	case *Caps:
		caps := value.(*Caps)
		C.X_gst_g_object_set_caps(e.GstElement, cname, caps.caps)
	case *Structure:
		structure := value.(*Structure)
		C.X_gst_g_object_set_structure(e.GstElement, cname, structure.C)
	}
}

func (e *Element) cleanCallback() {

	if e.onPadAdded == nil {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	delete(callbackStore, e.callbackID)
}

//export go_callback_new_pad
func go_callback_new_pad(CgstElement *C.GstElement, CgstPad *C.GstPad, callbackID C.guint64) {

	mutex.Lock()
	element := callbackStore[uint64(callbackID)]
	mutex.Unlock()

	if element == nil {
		return
	}

	callback := element.onPadAdded

	pad := &Pad{
		pad: CgstPad,
	}

	callback(element, pad)
}

func (e *Element) SetPadAddedCallback(callback PadAddedCallback) {
	e.onPadAdded = callback

	var callbackID uint64
	mutex.Lock()
	for {
		callbackID = rand.Uint64()
		if callbackStore[callbackID] != nil {
			continue
		}
		callbackStore[callbackID] = e
		break
	}
	mutex.Unlock()

	e.callbackID = callbackID

	detailedSignal := (*C.gchar)(C.CString("pad-added"))
	defer C.free(unsafe.Pointer(detailedSignal))

	runtime.SetFinalizer(e, func(e *Element) {
		e.cleanCallback()
	})

	C.X_g_signal_connect(e.GstElement, detailedSignal, C.guint64(callbackID))
}

func ElementFactoryMake(factoryName string, name string) (e *Element, err error) {
	var pName *C.gchar

	pFactoryName := (*C.gchar)(unsafe.Pointer(C.CString(factoryName)))
	defer C.g_free(C.gpointer(unsafe.Pointer(pFactoryName)))
	if name == "" {
		pName = nil
	} else {
		pName = (*C.gchar)(unsafe.Pointer(C.CString(name)))
		defer C.g_free(C.gpointer(unsafe.Pointer(pName)))
	}
	gstElt := C.gst_element_factory_make(pFactoryName, pName)

	if gstElt == nil {
		err = errors.New(fmt.Sprintf("could not create a GStreamer element factoryName %s, name %s", factoryName, name))
		return
	}

	e = &Element{
		GstElement: gstElt,
	}

	return
}

func nonCopyGoBytes(ptr uintptr, length int) []byte {
	var slice []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	header.Cap = length
	header.Len = length
	header.Data = ptr
	return slice
}

func nonCopyCString(data *C.char, size C.int) []byte {
	return nonCopyGoBytes(uintptr(unsafe.Pointer(data)), int(size))
}

func (e *Element) SetRtpBinNewStorageSignal(latency int) {

	C.X_g_signal_connect_rtpbin_newstorage(e.GstElement, C.guint64(latency))
}

func (e *Element) SetRtpBinFecDecSignal(pt int) {
	C.X_g_signal_connect_rtpbin_requestfecdec(e.GstElement, C.gint(pt))
}

func (e *Element) SetRtpBinRequestptmapSignal() {
	C.X_g_signal_connect_rtpbin_requestptmap(e.GstElement)
}

func (e *Element) SetRtpBinNewRtpJitterSignal(params string)  {
	n := (*C.gchar)(unsafe.Pointer(C.CString(params)))
	C.X_g_signal_connect_rtpbin_newjitterbuffer(e.GstElement,n)
}

