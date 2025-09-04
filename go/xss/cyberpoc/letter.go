package cyberpoc

import (
	"html/template"
	"sync"
)

type Letter struct {
	Mail    string
	Title   string
	Content template.HTML
}

type LetterRepo struct {
	letters []Letter
	mutex   sync.Mutex
}

func (r *LetterRepo) Save(letter Letter) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.letters = append(r.letters, letter)
}

func (r *LetterRepo) List() []Letter {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.letters
}
