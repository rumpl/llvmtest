package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"tinygo.org/x/go-llvm"
)

func main() {
	llvm.InitializeNativeTarget()
	llvm.InitializeNativeAsmPrinter()
	llvm.InitializeAllAsmParsers()

	target, _ := llvm.GetTargetFromTriple(llvm.DefaultTargetTriple())

	tm := target.CreateTargetMachine(
		llvm.DefaultTargetTriple(),
		"",
		"",
		llvm.CodeGenLevelNone,
		llvm.RelocDefault,
		llvm.CodeModelDefault,
	)
	passManager := llvm.NewPassManager()
	passManager.AddCFGSimplificationPass()
	passManager.AddConstantMergePass()
	passManager.AddGVNPass()
	passManager.AddReassociatePass()

	builder := llvm.NewBuilder()
	mod := llvm.NewModule("main")

	main := llvm.FunctionType(llvm.Int32Type(), []llvm.Type{}, false)
	llvm.AddFunction(mod, "main", main)
	block := llvm.AddBasicBlock(mod.NamedFunction("main"), "entry")
	builder.SetInsertPoint(block, block.FirstInstruction())

	fnType := llvm.FunctionType(llvm.Int32Type(), []llvm.Type{
		llvm.PointerType(llvm.Int8Type(), 0),
	}, true)

	llvm.AddFunction(mod, "printf", fnType)

	p := builder.CreateGlobalStringPtr("hello\n", "")
	builder.CreateCall(mod.NamedFunction("printf"), []llvm.Value{p}, "printf")

	p1 := builder.CreateGlobalStringPtr("world\n", "")
	builder.CreateCall(mod.NamedFunction("printf"), []llvm.Value{p1}, "printf")

	r := builder.CreateAlloca(llvm.Int32Type(), "")
	builder.CreateStore(llvm.ConstInt(llvm.Int32Type(), 0, false), r)
	rVal := builder.CreateLoad(r, "")
	builder.CreateRet(rVal)

	if ok := llvm.VerifyModule(mod, llvm.PrintMessageAction); ok != nil {
		fmt.Println(ok.Error())
		return
	}

	passManager.Run(mod)

	mod.Dump()

	passManager.Dispose()

	llvmBuf, _ := tm.EmitToMemoryBuffer(mod, llvm.ObjectFile)
	_ = ioutil.WriteFile("out.o", llvmBuf.Bytes(), 0666)

	exec.Command("cc", "out.o", "-fno-PIE", "-lc", "-o", "out").Run()
}
