package main

func acl_entry(id int) *ACLEntry {
	for ix, entry := range config.Users {
		if entry.Id == id {
			return &config.Users[ix]
		}
	}
	return nil
}

func acl_exists(id int) bool {
	return acl_entry(id) != nil
}

func acl_can(id int, ability string) bool {
	if ability == ACL_ANY {
		return true
	}
	if isMaster(id) {
		return true
	}
	e := acl_entry(id)
	if e == nil {
		return false
	}
	for _, s := range e.Allow {
		if ability == s {
			return true
		}
	}
	return false
}

func acl_all(abilities []string) (ids []int) {
	for _, entry := range config.Users {
		if len(abilities) == 0 {
			ids = append(ids, entry.Id)
		} else {
			able := false
		scan:
			for _, ability := range abilities {
				if ability == ACL_ANY {
					able = true
					break scan
				}
				if isMaster(entry.Id) {
					able = true
					break scan
				}
				for _, right := range entry.Allow {
					if right == ability {
						able = true
						break scan
					}
				}
			}
			if able {
				ids = append(ids, entry.Id)
			}
		}
	}
	return
}

func acl_touch(id int, name string) {
	dirty := false
	e := acl_entry(id)
	if e == nil {
		ee := ACLEntry{id, name, []string{}}
		config.Users = append(config.Users, ee)
		dirty = true
	} else if e.Name != name {
		e.Name = name
		dirty = true
	}
	if dirty {
		SaveConfig()
	}
}

func acl_grant(id int, name string) bool {
	if isMaster(id) {
		return false
	}
	if name == ACL_ANY {
		return false
	}
	if name == ACL_ALL {
		r := false
		for _, ability := range acl_abilities() {
			r = r || acl_grant(id, ability)
		}
		return r
	}
	if acl_can(id, name) {
		return false
	}
	e := acl_entry(id)
	e.Allow = append(e.Allow, name)
	return true
}

func acl_revoke(id int, name string) bool {
	if isMaster(id) {
		return false
	}
	e := acl_entry(id)
	if name == ACL_ALL {
		if len(e.Allow) == 0 {
			return false
		}
		e.Allow = []string{}
		return true
	}
	for ix, ability := range e.Allow {
		if ability == name {
			e.Allow = append(e.Allow[:ix], e.Allow[ix+1:]...)
			return true
		}
	}
	return false
}

func acl_abilities() (list []string) {
	c := make(map[string]int)
	for _, h := range handlers {
		if h.perm != ACL_ANY {
			c[h.perm] = 1
		}
	}
	c[ACL_INFORM] = 1
	c[ACL_SUPERVISE] = 1
	for k := range c {
		list = append(list, k)
	}
	return
}
