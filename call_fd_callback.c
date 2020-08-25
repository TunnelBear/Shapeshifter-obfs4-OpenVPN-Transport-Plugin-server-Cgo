#include <stdint.h>

typedef void (*fd_callback)(void *class_ptr, int);

// bridge to allow a function pointer to be passed to go. That ptr must be called from c.
void call_fd_callback(void* func_ptr, void* class_ptr, uintptr_t fd)
{
  fd_callback fd_func = (fd_callback) func_ptr;
  fd_func(class_ptr, fd);
}
