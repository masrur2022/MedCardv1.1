package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"medcard-new/begening/controllers/bycrypt"
	"medcard-new/begening/controllers/handlefile"
	"medcard-new/begening/controllers/jwtgen"
	"medcard-new/begening/controllers/velidation"
	"medcard-new/begening/structures"
	"os"

	// "strings"

	// "strings"

	// "os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GlobeStruct struct{
	QuestionId string `json:"questionid" bson:"_id"`
	QuestionsText string  `json:"questiontext"`
	QuestionsTitle string `json:"questiontitle"`
	QuestionsAuthorName string `json:"questionauthorname"`
}

var (
	ctx context.Context
	client *mongo.Client
)

var redirect_url string = os.Getenv("URL")
var DB_Url string = os.Getenv("DBURL")
func Authenticationservice(){
	if(DB_Url ==  ""){
		DB_Url = "mongodb://127.0.0.1:27017"
	}
	log.Printf("DB URl%v\n",DB_Url)
	log.Printf("DB URl%v\n",os.Getenv("DBURL"))
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")
	// clientOptions := options.Client().ApplyURI(os.Getenv("DBURL"))
	clientG, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Mongo.connect() ERROR: ", err)
	}
	ctxG, _ := context.WithTimeout(context.Background(), 15*time.Minute)
	// collection := client.Database("MedCard").Collection("users")
	ctx = ctxG
	client = clientG
}
func InsertQuestions(c *gin.Context){
	var(
		Questions GlobeStruct
		QuestionsDb GlobeStruct
	)
	c.ShouldBindJSON(&Questions)
	log.Printf("qjson: %v\n",Questions)
	// """"""""""""""""""""""JWT VALIDATION""""""""""""""""""""""""""
	ClaimsData := jwtgen.Velidation(c)
	log.Println(ClaimsData)
	// """"""""""""""""""""""JWT VALIDATION""""""""""""""""""""""""""
	Authenticationservice()
	collection := client.Database("MedCard").Collection("questions")
	collection.FindOne(ctx, bson.M{"questionstitle":Questions.QuestionsTitle}).Decode(&QuestionsDb)


	valueStruct,err := json.Marshal(Questions)
	if err != nil{
		log.Printf("Marshel Error: %v\n",err)
	}
	
	checkPointOne,checkPointTwo := velidation.TestTheStruct(c,"questiontitle:questiontext:questionauthorname",string(valueStruct),"FieldsCheck:true,DBCheck:false","","")
	if checkPointOne != false && checkPointTwo != true{
		var primitiveId = primitive.NewObjectID().Hex()
		Questions.QuestionId = primitiveId
	
		collection.InsertOne(ctx,Questions)
	}else{
		c.JSON(400, gin.H{
			"Code":"The Question Already exist",
		})
	}
}
func ProfileChange(c *gin.Context){
	var(
		CheckpointPassed bool = false
		ChangeStruct structures.GlobeStruct
		DecodedSigninStruct structures.SignupDoctor
	)
	jsonFM := c.Request.FormValue("json")
	files, handler, errIMG := c.Request.FormFile("img")
	// """""""""""""""""""""""check The file on existense"""""""""""""""""""""""
	// if errIMG != nil {
	// 	c.JSON(409, gin.H{
	// 		"sttus": "NOIMGFILEEXIST",
	// 	})
	// }

	files.Seek(23, 23)
	log.Printf("File Name %s\n", handler.Filename)
	// """""""""""""""""""""bind the request data into structure"""""""""""""""""""""
	json.Unmarshal([]byte(jsonFM), &ChangeStruct)
	// ChangeStruct.ImgUrl = handlefile.Handlefile(c,"../../static/upload")

	// """""""""""""""""""""bind the request data into structure"""""""""""""""""""""
	json.Unmarshal([]byte(jsonFM),&ChangeStruct)
	valueStruct ,err := json.Marshal(ChangeStruct)
	if err != nil{
		log.Printf("err%v",err)
	}

	CookieData := jwtgen.Velidation(c)
	Authenticationservice()
	collection := client.Database("MedCard").Collection("users")
	// collection.FindOne(ctx,)
	collection.FindOne(ctx,bson.M{"name":ChangeStruct.Name,"surname":ChangeStruct.Surname,"permissions":CookieData.Permissions}).Decode(&DecodedSigninStruct)


	if CookieData.Permissions == "admin"{
		var ChangeStruct structures.Admin
		json.Unmarshal([]byte(jsonFM), &ChangeStruct)
		checkPointOne,checkPointTwo := velidation.TestTheStruct(c,"phone:password:email:name:surname:lastname",string(valueStruct),"FieldsCheck:true,DBCheck:true","admin",CookieData.Id)
		log.Println(CheckpointPassed)
		if checkPointOne != false && checkPointTwo == false{
			log.Printf("ds1%v\n",ChangeStruct)
			ChangeStruct.Userid = CookieData.Id
			ChangeStruct.Permissions = CookieData.Permissions
			ChangeStruct.Password = DecodedSigninStruct.Password
			if errIMG != nil{
				ChangeStruct.ImgUrl = DecodedSigninStruct.ImgUrl
			}else{
				ChangeStruct.ImgUrl = handlefile.Handlefile(c,"../static/upload")
			}
			_ ,err := collection.ReplaceOne(ctx,bson.M{"_id":CookieData.Id},ChangeStruct)
			if err != nil{
				log.Printf("Err insert",err)
			}
			collection.InsertOne(ctx,ChangeStruct)
			c.JSON(200, gin.H{
				"Code":"Your Request Successfully Handeled",
			})
		}
	}else if CookieData.Permissions == "doctor"{
		var ChangeStruct structures.SignupDoctor
		json.Unmarshal([]byte(jsonFM), &ChangeStruct)
		checkPointOne,checkPointTwo := velidation.TestTheStruct(c,"phone:password:name:surname:lastname:position:adress:biography:email",string(valueStruct),"FieldsCheck:true,DBCheck:true","doctor",CookieData.Id)
		log.Println(CheckpointPassed)
		if checkPointOne != false && checkPointTwo == false{
			log.Printf("ds1%v\n",ChangeStruct)
			ChangeStruct.Userid = CookieData.Id
			ChangeStruct.Permissions = CookieData.Permissions
			ChangeStruct.Password = DecodedSigninStruct.Password
			if errIMG == nil{
				log.Printf("123%v\n",ChangeStruct)
				ChangeStruct.ImgUrl = DecodedSigninStruct.ImgUrl
			}else{
				log.Printf("456%v\n",ChangeStruct)
				ChangeStruct.ImgUrl = handlefile.Handlefile(c,"../static/upload")
			}
			ChangeStruct.History = append(ChangeStruct.History, structures.History{
				Year: "2022-12",
				Position: "jfdfdd",
			})
			_ ,err := collection.ReplaceOne(ctx,bson.M{"_id":CookieData.Id},ChangeStruct)
			if err != nil{
				log.Printf("Err insert",err)
			}
			collection.InsertOne(ctx,ChangeStruct)
			c.JSON(200, gin.H{
				"Code":"Your Request Successfully Handeled",
			})
		}
	}else if CookieData.Permissions == "client"{
		var ChangeStruct structures.Signup
		var DecodedSigninStruct structures.Signup
		json.Unmarshal([]byte(jsonFM), &ChangeStruct)
		checkPointOne,checkPointTwo := velidation.TestTheStruct(c,"email:phone:password",string(valueStruct),"FieldsCheck:true,DBCheck:true","client",CookieData.Id)
		fmt.Printf("checkPointOne: %v\n", checkPointOne)
		fmt.Printf("checkPointTwo: %v\n", checkPointTwo)
		collection.FindOne(ctx,bson.M{"name":ChangeStruct.Name,"surname":ChangeStruct.Surname,"permissions":CookieData.Permissions}).Decode(&DecodedSigninStruct)
		if checkPointOne != false && checkPointTwo == false{
			hashedPass ,err := bycrypt.HashPassword(DecodedSigninStruct.Password)
				if err != nil{
				log.Printf("Err Hash%v",err)
			}
			log.Printf("ds1%v\n",ChangeStruct)
			DecodedSigninStruct.Userid = CookieData.Id
			DecodedSigninStruct.Permissions = CookieData.Permissions
			DecodedSigninStruct.Password = hashedPass
			if errIMG != nil{
				log.Printf("123%v\n",ChangeStruct)
				// ChangeStruct.ImgUrl = DecodedSigninStruct.ImgUrl
			}else{
				log.Printf("456%v\n",ChangeStruct)
				DecodedSigninStruct.ImgUrl = handlefile.Handlefile(c,"../static/upload")
			}
			fmt.Printf("DecodedSigninStruct: %v\n", DecodedSigninStruct)
			_ ,err = collection.ReplaceOne(ctx,bson.M{"_id":CookieData.Id},DecodedSigninStruct)
			if err != nil{
				log.Printf("Err insert",err)
			}
			// log.Printf("cur",cur)
			c.JSON(200, gin.H{
				"Code":"Your Request Successfully Handeled",
			})
		}
	}
}

func Cors(c *gin.Context) {
	if(redirect_url == ""){
		redirect_url = "http://127.0.0.1:5502"
	}
	log.Printf("url%v\n",redirect_url)
	c.Writer.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, ResponseType, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}

	c.Next()
}