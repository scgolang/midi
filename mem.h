#ifndef MEM_H_INCLUDED
#define MEM_H_INCLUDED

extern void *midi_mem_alloc(long nbytes, const char * file, int line);
extern void *midi_mem_calloc(int count, long nbytes, const char * file, int line);
extern void  midi_mem_free(void *ptr, const char *file, int line);
extern void *midi_mem_resize(void *ptr, long nbytes, const char *file, int line);

#define ALLOC(nbytes)         midi_mem_alloc((nbytes), __FILE__, __LINE__)
#define CALLOC(count, nbytes) midi_mem_calloc((count), (nbytes), __FILE__, __LINE__)

#define NEW(p) ((p) = ALLOC((long) sizeof *(p)))
#define NEW0(p) ((p) = CALLOC(1, (long) sizeof *(p)))
#define FREE(p) ((void) (midi_mem_free((p), __FILE__, __LINE__), (p) = 0))
#define RESIZE(p, nbytes) ((p) = midi_mem_resize((p), (nbytes), __FILE__, __LINE__))

#endif // MEM_H_INCLUDED
