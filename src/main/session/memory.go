package session

import (
    "time"
)

type MemorySessioner struct {
    accessTime int64
    id string
    values map[interface{}]interface{}
}

func (s *MemorySessioner) Set(key, value interface{}) error {
    s.values[key] = value
    return nil
}

func (s *MemorySessioner) Get(key interface{}) interface{} {
    if val, ok := s.values[key]; ok {
        return val
    }
    return nil
}

func (s *MemorySessioner) Delete(key interface{}) error {
    delete(s.values, key)
    return nil
}

func (s *MemorySessioner) SessionID() string {
    return s.id
}

func (s *MemorySessioner) Init(sid string) {
    s.id = sid
    s.values = make(map[interface{}]interface{})
}

func (s *MemorySessioner) GetAccessTime() int64 {
    return s.accessTime
}

func (s *MemorySessioner) SetAccessTime(accessTime int64) {
    s.accessTime = accessTime
}

type MemoryProvider struct {
    sessionMap map[string]*MemorySessioner
}

func (s *MemoryProvider) Init() {
	s.sessionMap = make(map[string]*MemorySessioner)
}

func (mp *MemoryProvider) SessionInit(sid string) (Session, error) {
    if _, ok := mp.sessionMap[sid]; ok {
        mp.sessionMap[sid].SetAccessTime(time.Now().Unix())
        return mp.sessionMap[sid], nil
    } else {
        mp.sessionMap[sid] = new(MemorySessioner)
        mp.sessionMap[sid].Init(sid)
        mp.sessionMap[sid].SetAccessTime(time.Now().Unix())
        return mp.sessionMap[sid], nil
    }
}

func (mp *MemoryProvider) SessionRead(sid string) (Session, error) {
    if session, ok := mp.sessionMap[sid]; ok {
        session.SetAccessTime(time.Now().Unix())
        return mp.sessionMap[sid], nil
    } else {
        mp.sessionMap[sid] = new(MemorySessioner)
        mp.sessionMap[sid].Init(sid)
        mp.sessionMap[sid].SetAccessTime(time.Now().Unix())
        return mp.sessionMap[sid], nil
    }
}

func (mp *MemoryProvider) SessionDestroy(sid string) error {
    delete((*mp).sessionMap, sid)
    return nil
}

func (mp *MemoryProvider) SessionGC(maxLifeTime int64) {
    for sid, session := range (*mp).sessionMap {
        if session.GetAccessTime() < (time.Now().Unix() - maxLifeTime) {
            delete((*mp).sessionMap, sid)
        }
    }
}
