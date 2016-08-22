package lowlevel

type LV struct {
	Name string `mapstructure:"LVM2_LV_NAME"`
}

func QueryLV(vg string) (map[string]*LV, error) {
	result := map[string]*LV{}
	lvs, err := queryLVM("lvs", "all", []string{})
	if err != nil {
		return nil, err
	}
	for _, values := range lvs {
		lv := &LV{}
		if err = parseLVM(values, lv); err != nil {
			return nil, err
		}
		result[lv.Name] = lv
	}
	return result, nil
}
