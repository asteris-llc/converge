package lowlevel

type LogicalVolume struct {
	Name string `mapstructure:"LVM2_LV_NAME"`
}

func QueryLogicalVolumes(vg string) (map[string]*LogicalVolume, error) {
	result := map[string]*LogicalVolume{}
	lvs, err := queryLVM("lvs", "all", []string{})
	if err != nil {
		return nil, err
	}
	for _, values := range lvs {
		lv := &LogicalVolume{}
		if err = parseLVM(values, lv); err != nil {
			return nil, err
		}
		result[lv.Name] = lv
	}
	return result, nil
}
