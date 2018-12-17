package model

type TWapPayModel struct {
}

var Wap TWapPayModel

func (t *TWapPayModel)GetMchParams(mid string,paytype int) (map[string]string,error)  {
	params := make(map[string]string)
	var name,val string
	rows,err := gdb.Raw("select attrname,attrvalue from ap_mchpay_options where mid = ? and paytype = ?",mid,paytype).Rows()
	if err != nil {
		return nil,err
	}
	defer rows.Close()
	for rows.Next()  {
		rows.Scan(&name,&val)
		params[name] = val
	}
	return params,nil
}
