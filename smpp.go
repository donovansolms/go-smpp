package smpp34

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type Smpp struct {
	mu        sync.Mutex
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	Connected bool
	Sequence  uint32
	Bound     bool
}

type Params map[string]interface{}

func NewSmppConnect(host string, port int) (*Smpp, error) {
	s := &Smpp{}

	err := s.Connect(host, port)

	return s, err
}

func (s *Smpp) Connect(host string, port int) (err error) {
	s.conn, err = net.Dial("tcp", host+":"+strconv.Itoa(port))

	return err
}

func (s *Smpp) NewSeqNum() uint32 {
	defer s.mu.Unlock()

	s.mu.Lock()
	s.Sequence++
	return s.Sequence
}

func (s *Smpp) Bind(system_id string, password string, params *Params) (Pdu, error) {
	b, _ := NewBind(
		&Header{Id: BIND_TRANSCEIVER},
		[]byte{},
	)

	b.SetField(INTERFACE_VERSION, 0x34)
	b.SetField(SYSTEM_ID, system_id)
	b.SetField(PASSWORD, password)
	b.SetSeqNum(s.NewSeqNum())

	for f, v := range *params {
		err := b.SetField(f, v)

		if err != nil {
			return nil, err
		}
	}

	return Pdu(b), nil
}

func (s *Smpp) EnquireLink() (Pdu, error) {
	p, _ := NewEnquireLink(
		&Header{
			Id:       ENQUIRE_LINK,
			Sequence: s.NewSeqNum(),
		},
	)

	return Pdu(p), nil
}

func (s *Smpp) EnquireLinkResp(seq uint32) (Pdu, error) {
	p, _ := NewEnquireLinkResp(
		&Header{
			Id:       ENQUIRE_LINK_RESP,
			Status:   ESME_ROK,
			Sequence: seq,
		},
	)

	return Pdu(p), nil
}

func (s *Smpp) SubmitSm(source_addr, destination_addr, short_message string, params *Params) (Pdu, error) {

	p, _ := NewSubmitSm(
		&Header{
			Id:       SUBMIT_SM,
			Sequence: s.NewSeqNum(),
		},
		[]byte{},
	)

	p.SetField(SOURCE_ADDR, source_addr)
	p.SetField(DESTINATION_ADDR, destination_addr)
	p.SetField(SHORT_MESSAGE, short_message)

	for f, v := range *params {
		err := p.SetField(f, v)

		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (s *Smpp) Unbind() (Pdu, error) {
	p, _ := NewUnbind(
		&Header{
			Id:       UNBIND,
			Sequence: s.NewSeqNum(),
		},
	)

	return Pdu(p), nil
}

func (s *Smpp) Read() (Pdu, error) {
	l := make([]byte, 4)
	_, err := s.conn.Read(l)
	if err != nil {
		return nil, err
	}

	pduLength := unpackUi32(l) - 4
	if pduLength > MAX_PDU_SIZE {
		return nil, errors.New("PDU Len larger than MAX_PDU_SIZE")
	}

	data := make([]byte, pduLength)

	i, err := s.conn.Read(data)
	if err != nil {
		return nil, err
	}

	if i != int(pduLength) {
		return nil, errors.New("PDU Len different than read bytes")
	}

	pkt := append(l, data...)
	fmt.Println(hex.Dump(pkt))

	pdu, err := ParsePdu(pkt)
	if err != nil {
		return nil, err
	}

	return pdu, nil
}

func (s *Smpp) Write(p Pdu) error {
	_, err := s.conn.Write(p.Writer())

	fmt.Println(hex.Dump(p.Writer()))

	return err
}

func (s *Smpp) Close() {
	s.conn.Close()
}