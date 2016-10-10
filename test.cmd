go install icepacker\cmd\ipack

echo Pack and Unpack test1
ipack pack r:\icepack-test\test1 r:\icepack-test\test1.pack
ipack list r:\icepack-test\test1.pack
ipack unpack r:\icepack-test\test1.pack r:\icepack-test\unpacked\test1

echo Pack and Unpack test2
ipack pack -c gzip r:\icepack-test\test2 r:\icepack-test\test2.pack
ipack list r:\icepack-test\test2.pack
ipack unpack r:\icepack-test\test2.pack r:\icepack-test\unpacked\test2

echo Pack and Unpack test3
ipack pack -e aes --key secretkey r:\icepack-test\test3 r:\icepack-test\test3.pack
ipack list --key secretkey r:\icepack-test\test3.pack
ipack unpack --key secretkey r:\icepack-test\test3.pack r:\icepack-test\unpacked\test3

echo Pack and Unpack filetest
ipack pack -c gzip -e aes --key secretkey r:\icepack-test\filetest r:\icepack-test\filetest.pack
ipack list --key secretkey r:\icepack-test\filetest.pack
ipack unpack --key secretkey r:\icepack-test\filetest.pack r:\icepack-test\unpacked\filetest

echo Pack and Unpack merged file
ipack pack r:\icepack-test\test1 r:\icepack-test\merge.pack
copy /b r:\icepack-test\bb.jpg + r:\icepack-test\merge.pack r:\icepack-test\merged.pack.jpg
ipack unpack r:\icepack-test\merged.pack.jpg r:\icepack-test\unpacked\merged
