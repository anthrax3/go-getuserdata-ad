// Copyright 2015 
// Выгрузка даных из

// File contains a modify example
package main

import (
	"fmt"
	"igo-ldap"   // https://github.com/go-ldap/ldap/
	"strings"	
	"os"
)

//тип данных пользователя
type UserData struct {
	name string  // имя пользователя
	tel string  // внутр номер теелфона
	manager string  // рукводитель
	department string  // отдел
}

type CfgServer struct{
	server  string  // сервер контроллера домена
	users  string  // пользователь который имеет доступ к AD
	searchs string // путь до OU в котором будет проходить поиск
	company string // название компании (филиала) как указано в AD
	department string  // название отдела компании (филиала) по которому нужно искать учетки

}


func Savestrtofile(namef string, str string) int {
	file, err := os.Create(namef)
	if err != nil {
		// handle the error here
		return -1
	}
	defer file.Close()

	file.WriteString(str)
	return 0
}

// сортировка пузырьком
func sort_bubbles(slice []UserData) {
    var length int = len(slice) - 1

	for i := 0; i < len(slice); i++ {
		for j := 0; j < length; j++ {
			if slice[j].manager > slice[j+1].manager {
				slice[j], slice[j+1] = slice[j+1], slice[j]
			}
		}
		length--
	}
	
}

//выделение ФИО руководителя 
func getManager(s string) string {
	return ((strings.Split(s,","))[0])[3:]
}


// получение данных из AD
func GetListUserfromAD(users string, pass string,server string, ssearch string,company string, department string) []UserData {
	var res []UserData
	var tmp UserData
	res = make([]UserData,0)
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", server, 389))
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()

	err = l.Bind(users, pass)
	if err != nil {
		fmt.Println(err)
	}
	
	searchRequest := ldap.NewSearchRequest(
		ssearch, // The base dn to search    
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalPerson))", // The filter to apply  		
		[]string{"dn", "cn","company","department","description","mail","telephoneNumber","manager","name"},                    // A list attributes to retrieve
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		fmt.Println(err)
	}
	
	for _, entry := range sr.Entries {
		scompany := entry.GetAttributeValue("company")
		sotdel := entry.GetAttributeValue("department")
		if (strings.Compare(scompany,company)==0) && (strings.Compare(sotdel,department)==0) {
			tmp.name = entry.GetAttributeValue("cn")
			tmp.manager = getManager(entry.GetAttributeValue("manager"))
			tmp.tel = entry.GetAttributeValue("telephoneNumber")
			tmp.department = entry.GetAttributeValue("department")
			res = append(res,tmp)
		}		
	}
	return res
}


//---------------------------

// чтение из текстового конфиг файла config.cfg для конфигурации сервера и поиска
func (ss  *CfgServer) Readcfg(namef string) {	
	str := readfiletxt(namef)
	s := strings.Split(str, ";")
	ss.server=s[0]
	ss.users=s[1]
	ss.searchs=s[2]
	ss.company=s[3]
	ss.department=s[4]
}

// чтение из текстового конфиг файла list-rg.cfg список руководителей по которым будут
// разделены точкой с запятой, например, ФИО1;ФИО2;
func ReadList(namef string) []string {	
    ls:=make( []string,0)
	str := readfiletxt(namef)
	s := strings.Split(str, ";")
    for i:=0;i<len(s);i++ {
		if s[i]!="" {
			ls=append(ls,s[i])
		}
	}
	return ls
}

//// чтение файла с именем namefи возвращение содержимое файла, иначе текст ошибки
func readfiletxt(namef string) string {
	file, err := os.Open(namef)
	if err != nil {
		return "handle the error here"
	}
	defer file.Close()
	// get the file size
	stat, err := file.Stat()
	if err != nil {
		return "error here"
	}
	// read the file
	bs := make([]byte, stat.Size())
	_, err = file.Read(bs)
	if err != nil {
		return "error here"
	}
	return string(bs)
}

//---------------------------


func main() {
	var userarray []UserData
	var passw string
	var str string
	var c CfgServer
	
	c.Readcfg("config-ad.cfg")
	
	fmt.Println(c)
	
	// список руководителей по которым нужно выгрузить подчиненных
	list_rg:= ReadList("list-rg.cfg")

	fmt.Println("Введите пароль для доступа к AD: ")
    fmt.Scanf("%s", &passw)
	
	userarray = GetListUserfromAD(c.users,passw,c.server,c.searchs,c.company,c.department)	
	sort_bubbles(userarray)
	
	for i:=0;i<len(userarray);i++{
		for j:=0;j<len(list_rg);j++ {
			if userarray[i].manager==list_rg[j] {	
				str += userarray[i].tel+";"+userarray[i].name+";"+userarray[i].manager+"\n"	
				break
			}
		}
	}
		
	
	Savestrtofile("list-num-tel.cfg",str)	
}
 