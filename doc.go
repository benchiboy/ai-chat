// ai-chat project doc.go

/*
ai-chat document
*/
package main

/*

//				var header MsgHeader
//				json.Unmarshal([]byte(string(client.send)), &header)
//				switch header.MsgType {
//				case MSG_USER_ID:
//					fmt.Println("GET USER ID")
//					var getUser MsgGetUserId
//					getUser.FromUserId = client.from_user_id
//					getUser.MsgType = header.MsgType
//					msgBuf, _ := json.Marshal(getUser)
//					client.send = msgBuf
//					h.write_user <- client
//				case MSG_TEXT:
//					h.write_user <- client
//					c := h.getClient(client.from_user_id, client.to_user_id)
//					if c == nil {
//						r := new(Client)
//						r.conn = client.conn
//						r.group = client.group
//						r.send = []byte("")
//						h.write_user <- r
//					} else {
//						var textMsg MsgText
//						json.Unmarshal([]byte(string(client.send)), &textMsg)
//						textMsg.ToUserId = c.from_user_id
//						textMsg.FromUserId = client.from_user_id
//						msgBuf, _ := json.Marshal(textMsg)
//						client.to_user_id = c.from_user_id
//						c.to_user_id = client.from_user_id
//						c.send = msgBuf
//						h.write_user <- c
//					}
//				}
*/
