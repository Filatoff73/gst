package gst

/*
#cgo pkg-config: gstreamer-1.0
#include "gst.h"
*/
import "C"

import (
	"runtime"
)

type Bus struct {
	C *C.GstBus
}

func (b *Bus) Pop() (message *Message) {

	CGstMessage := C.gst_bus_pop(b.C)
	if CGstMessage == nil {
		return nil
	}
	message = &Message{
		C: CGstMessage,
	}

	runtime.SetFinalizer(message, func(message *Message) {
		C.gst_message_unref(message.C)
	})

	return
}

func (b *Bus) PopFiltered(messageType MessageType) (message *Message) {

	CGstMessage := C.gst_bus_pop_filtered(b.C, C.GstMessageType(messageType))
	if CGstMessage == nil {
		return nil
	}
	message = &Message{
		C: CGstMessage,
	}

	runtime.SetFinalizer(message, func(message *Message) {
		C.gst_message_unref(message.C)
	})

	return
}

func (b *Bus) Pull(messageType MessageType) (message *Message) {

	CGstMessage := C.gst_bus_poll(b.C, C.GstMessageType(messageType), 18446744073709551615)
	if CGstMessage == nil {
		return nil
	}

	message = &Message{
		C: CGstMessage,
	}

	runtime.SetFinalizer(message, func(message *Message) {
		C.gst_message_unref(message.C)
	})

	return
}

func (b *Bus) HavePending() bool {
	return C.gst_bus_have_pending(b.C) != 0
}
