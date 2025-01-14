package phone

import (
	"errors"
	"fmt"
	"github.com/totoval/framework/helpers/log"
	"github.com/totoval/framework/helpers/toto"
	"totoval/app/logics/phone/interfaces"
	"totoval/app/logics/phone/messages"
	"totoval/app/logics/phone/sms"
)

type phone struct {
	chip interfaces.Chipper
	notifier interfaces.Notifier
	//smsBox interfaces.SmsBoxer
}

func New(chip interfaces.Chipper, notifier interfaces.Notifier) *phone {
	return &phone{
		chip: chip,
		notifier: notifier,
	}
}
func (ph *phone)Listen(){

	go func() {
		for{
			ph.chip.Read2()
		}
	}()


	for {
		//n, b, err := ph.chip.Read()

		select {
			case b :=<- ph.chip.Bytes():
				log.Info("Incoming data", toto.V{"data": string(b[:])})
				_ = log.Error(ph.parse(b))
			case err := <-ph.chip.Error():
				log.Panic(err)
		default:
			continue
		}

	}
}


func (ph *phone)parse(msg []byte) error {

	if messages.ParseOk(msg){
		return nil
	}

	// parse sms index +CMTI: "SM",2
	if matched, smsIndex, err := messages.ParseSmsIndex(msg); matched {
		if err != nil{
			return err
		}
		// sms receive event
		if err := sms.Retrieve(ph.chip, smsIndex); err != nil{
			return err
		}
		return nil
	}

	// parse sms content +CMGR: 1,\"\",145\r\n0891683108200905F5240FA101960119145036F90008918091110431237C3010006C00750063006B0069006E00200063006F006600660065006530116211731C4F6060F3559D70B94EC04E48FF0C90014F600035002E003562985168573A996E54C15238FF0C5168573A996E54C1768653EF4F7F7528FF5E53BB004100500050002F5C0F7A0B5E8F559D4E00676F002056DE0054004490008BA2
	if matched, pduStr, err := messages.ParseSmsContent(msg); matched {
		if err != nil{
			return err
		}

		smsContentPDU := NewPDU()
		smsContentPDU.Scan(pduStr)
		sender, content, err := smsContentPDU.Data()
		if err != nil{
			return err
		}

		return ph.notifier.Notify(sender, content)
	}

	//@todo other message type



	return errors.New(fmt.Sprintf("Not a normal message: %s", string(msg[:])))// not a valid
}

