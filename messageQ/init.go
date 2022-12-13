package messageQ

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		zap.L().Panic(msg,
			zap.Error(err),
		)
	}
}

// MQRead 包括序列化的chan，exchange名称，comsumer的Tag
func MQRead(messageChan chan<- []byte, exName string, qName string, comTag string) {
	conn, err := amqp.Dial("amqp://admin:csd@123@8.219.103.237:31656/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exName,   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		qName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// 配置Qos，投递到本地consumer队列上的最大消息数量
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		exName, // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		comTag, // consumer
		false,  // auto-ack设置为false，需要手动确认，消息才会从队列中删除
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs { //从msgs队列中取出消息（mq队列结构体）
			log.Printf(" [x] %s", d.Body)
			messageChan <- d.Body // 将消息发送到orm层写入数据库
			err := d.Ack(false)   // 成功写入后确认，每次只确认一个
			failOnError(err, "Failed to register a d")
		}
	}()

	zap.L().Info(exName + "Queue init successful, waiting for messages")
	log.Printf(" %s [*] Waiting for logs. To exit press CTRL+C", exName)
	<-forever
}
