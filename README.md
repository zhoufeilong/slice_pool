# slice_pool
decrease gc cost,self pool

如果你的程序因为内存中长驻大量对象导致gc 消耗cpu严重.可以考虑使用slice缓存数据减少gc开销了。
pool.go写了一个将数据同一管理的pool类。方便项目代码的书写。
如果你的使用频率低，可以直接使用[]TestStruct,以提高效率

如果你有疑问，或者我的代码有改善之处，或者对你有帮助，请給我一个回复。联系方式：619303503@qq.com
