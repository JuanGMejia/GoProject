package authentication

import (
	"crypto/rsa"
	"io/ioutil"
	"log"
	"src/github.com/dgrijalva/jwt-go"
	"HelloWorldGo/src/models"
	"time"
	"net/http"
	"encoding/json"
	"fmt"
	"src/github.com/dgrijalva/jwt-go/request"
)

//Se firman los tokens con una llave privada (openssl genrsa -out private.rsa 1024)
//se verifican con una llave publica (openssl rsa -in private.rsa -pubout > public.rsa.pub)

var (
	privateKey *rsa.PrivateKey
	publicKey *rsa.PublicKey
)

func init(){
	privateBytes, err := ioutil.ReadFile("E:/Repositorio/GoJuanG/GoProject/src/HelloWorldGo/src/keys/private.rsa")
	if err != nil {
		log.Fatal("No se pudo leer el archivo privado")
	}

	publicBytes, err := ioutil.ReadFile("E:/Repositorio/GoJuanG/GoProject/src/HelloWorldGo/src/keys/public.rsa.pub")
	if err != nil {
		log.Fatal("No se pudo leer el archivo publico")
	}

	privateKey,err = jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	if err !=nil{
		log.Fatal("No se pudo hacer el parse a private key")
	}

	publicKey,err = jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	if err !=nil{
		log.Fatal("No se pudo hacer el parse a public key")
	}
}

func GenerateJWT(user models.User)(string){
	claims := models.Claim{
		user,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour*1).Unix(),
			Issuer:"Taller de sabado",},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256,claims)
	result, err := token.SignedString(privateKey)
	if err !=nil{
		log.Fatal("No se pudo firmar el token")
	}
	return result
}

func Login(w http.ResponseWriter, r *http.Request){
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		fmt.Fprintln(w,"Error al leer usuario %s", err)
	}

	if user.Name== "juan" && user.Password=="juan"{
		user.Password=""
		user.Role = "admin"
		token := GenerateJWT(user)
		result := models.ResponseToken{token}
		jsonResult, err := json.Marshal(result)
		if err != nil{
			fmt.Fprintln(w,"Error al generar el json")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type","application/json")
		w.Write(jsonResult)
	}else {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w,"Usuario o clave no validos")
	}
}

func ValidateToken(w http.ResponseWriter, r *http.Request){
	token, err := request.ParseFromRequestWithClaims(r, request.OAuth2Extractor,&models.Claim{},func(token * jwt.Token)(interface{},error){
		return publicKey,nil
	})



	if err!=nil{
		switch err.(type) {
		case *jwt.ValidationError:
			vErr := err.(*jwt.ValidationError)
			switch vErr.Errors {
			case jwt.ValidationErrorExpired:
				fmt.Fprintln(w,"Su token ha expirado")
				return
			case jwt.ValidationErrorSignatureInvalid:
				fmt.Fprintln(w,"La firma del token no coincide")
				return
			default:
				fmt.Fprintln(w,"Su token no es valido")
				return

			}
		default:
			fmt.Fprintln(w,"Su token no es valido")
			return
		}
	}

	if token.Valid {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w,"Bienvenido al sistema")
	}else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w,"Su token no es valido")
	}

}
